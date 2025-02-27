package test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
	"github.com/samuel/prueba-orden/ordenJson"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// ~ ESTRUCTURAS PARA REGISTRO DE EVENTOS ~
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestLog struct {
	FechaTest     string         `json:"fecha_test"`
	NombreTest    string         `json:"nombre_test"`
	IngresoDatos  TestInput      `json:"ingreso_datos"`
	Procesos      []string       `json:"procesos"`
	ResultadosEsperados        ResultadosEsperados `json:"resultados_esperados"`
	ResultadosObtenidos        ResultadosObtenidos   `json:"resultados_obtenidos"`
	EstadoTest        string         `json:"estado_test"`
	TiempoDeEjecucion string         `json:"tiempo_ejecucion,omitempty"`
}

type TestInput struct {
	RawJSON string      `json:"raw_json,omitempty"`
	Parametros  interface{} `json:"parametros,omitempty"`
}

type ResultadosEsperados struct {
	ClavesOrdenadas []string    `json:"claves_ordenadas,omitempty"`
	TipoError   string      `json:"tipo_error,omitempty"`
	CustomCheck interface{} `json:"custom_check,omitempty"`
}

type ResultadosObtenidos struct {
	ClavesOrdenadas []string `json:"claves_ordenadas,omitempty"`
	JsonSalida  string   `json:"json_salida,omitempty"`
	Error       string   `json:"error,omitempty"`
}

type TestLogger struct {
	mu    sync.Mutex
	logs  map[string]*TestLog
}

var registradorGlobal = &TestLogger{
	logs: make(map[string]*TestLog),
}

func (tl *TestLogger) IniciadorTest(testName string, input interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	
	var rawJSON string
	switch v := input.(type) {
	case string:
		rawJSON = v
	case map[string]interface{}:
		if jsonStr, err := json.Marshal(v); err == nil {
			rawJSON = string(jsonStr)
		}
	}

	tl.logs[testName] = &TestLog{
		FechaTest: time.Now().Format(time.RFC3339Nano),
		NombreTest:  testName,
		IngresoDatos: TestInput{
			RawJSON: rawJSON,
			Parametros:  input,
		},
		Procesos: []string{"Test inicializado"},
		EstadoTest:   "En ejecución",
	}
}

func (tl *TestLogger) AgregarProceso(testName, process string) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	
	if logEntry, exists := tl.logs[testName]; exists {
		logEntry.Procesos = append(logEntry.Procesos, process)
	}
}

func (tl *TestLogger) GuardarResultado(testName string, actual ResultadosObtenidos, status string) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	
	if logEntry, exists := tl.logs[testName]; exists {
		logEntry.ResultadosObtenidos = actual
		logEntry.EstadoTest = status
		logEntry.TiempoDeEjecucion = time.Since(time.Now()).String()
	}
}

func (tl *TestLogger) ConfigResultadoEsperado(testName string, expected ResultadosEsperados) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	
	if logEntry, exists := tl.logs[testName]; exists {
		logEntry.ResultadosEsperados = expected
	}
}

func (tl *TestLogger) CrearArchivoLog() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	logEntries := make([]TestLog, 0, len(tl.logs))
	for _, entry := range tl.logs {
		logEntries = append(logEntries, *entry)
	}

	file, err := json.MarshalIndent(logEntries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("test_events.log", file, 0644)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// ~ CÓDIGO EXISTENTE CON REGISTRO DE EVENTOS ~
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type DocumentMetadata struct {
	TipoDocumento  string
	RUTCliente     string
	CmTitle        string
	Origen         string
}

var keyRegex = regexp.MustCompile(`"([^"]+)":`)

func extraerClavesJSON(orderedJSON string) []string {
	matches := keyRegex.FindAllStringSubmatch(orderedJSON, -1)
	keys := make([]string, 0, len(matches))
	for _, m := range matches {
		keys = append(keys, m[1])
	}
	return keys
}

func TestOrdenarJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "campos básicos ordenados",
			input: `{
				"cm:description": "desc",
				"tanner:rut-cliente": "123",
				"tanner:tipo-documento": "anexo",
				"cm:title": "title"
			}`,
			expected: []string{
				"tanner:tipo-documento",
				"tanner:rut-cliente",
				"cm:title",
				"cm:description",
			},
		},
		{
			name: "todos los campos presentes",
			input: `{
				"tanner:estado-visado": "aprobado",
				"tanner:tipo-documento": "contrato",
				"cm:versionLabel": "v1.0",
				"tanner:sub-categorias": "subcat"
			}`,
			expected: []string{
				"tanner:tipo-documento",
				"tanner:estado-visado",
				"tanner:sub-categorias",
				"cm:versionLabel",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testName := t.Name()
			startTime := time.Now()
			registradorGlobal.IniciadorTest(testName, tt.input)
			registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: tt.expected})

			registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON")
			got, err := ordenJson.OrdenarJSON(tt.input)

			var actual ResultadosObtenidos
			if err != nil {
				actual = ResultadosObtenidos{Error: err.Error()}
				registradorGlobal.GuardarResultado(testName, actual, "Fallido")
				t.Fatalf("OrdenarJSON() error = %v", err)
			}

			keys := extraerClavesJSON(got)
			actual = ResultadosObtenidos{
				ClavesOrdenadas: keys,
				JsonSalida:  got,
			}

			status := "Completado"
			if !reflect.DeepEqual(keys, tt.expected) {
				status = "Fallido"
				t.Errorf("Orden de claves incorrecto")
			}

			registradorGlobal.GuardarResultado(testName, actual, status)
			registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
		})
	}
}

func TestOrdenarDocumentoMetadata(t *testing.T) {
	metadata := ordenJson.DocumentMetadata{
		TipoDocumento:  "contrato",
		RUTCliente:     "12345678-9",
		CmTitle:        "Contrato de Servicios",
		Origen:         "Departamento Legal",
	}

	expectedOrder := []string{
		"tanner:tipo-documento",
		"tanner:rut-cliente",
		"tanner:origen",
		"cm:title",
	}

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, metadata)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: expectedOrder})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarDocumentoMetadata")
	orderedJSON, err := ordenJson.OrdenarDocumentoMetadata(metadata)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatalf("OrdenarDocumentoMetadata() error = %v", err)
	}

	keys := extraerClavesJSON(orderedJSON)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  orderedJSON,
	}

	status := "Completado"
	if !reflect.DeepEqual(keys, expectedOrder) {
		status = "Fallido"
		t.Errorf("Orden de claves incorrecto")
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestOrdenarMapaComoDocumentoMetadata(t *testing.T) {
	inputMap := map[string]interface{}{
		"cm:description":        "descripción",
		"tanner:rut-cliente":    "98765432-1",
		"tanner:tipo-documento": "informe",
		"tanner:categorias":     "legal",
	}

	expectedOrder := []string{
		"tanner:tipo-documento",
		"tanner:rut-cliente",
		"tanner:categorias",
		"cm:description",
	}

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, inputMap)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: expectedOrder})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarMapaComoDocumentoMetadata")
	orderedJSON, err := ordenJson.OrdenarMapaComoDocumentoMetadata(inputMap)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatalf("OrdenarMapaComoDocumentoMetadata() error = %v", err)
	}

	keys := extraerClavesJSON(orderedJSON)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  orderedJSON,
	}

	status := "Completado"
	if !reflect.DeepEqual(keys, expectedOrder) {
		status = "Fallido"
		t.Errorf("Orden de claves incorrecto")
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestOrdenarJSON_TipoDatoNoSoportado(t *testing.T) {
	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, 123)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{TipoError: "Tipo no soportado"})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con tipo no soportado")
	_, err := ordenJson.OrdenarJSON(123)

	var actual ResultadosObtenidos
	if err == nil {
		actual = ResultadosObtenidos{Error: "Se esperaba error para tipo no soportado, pero no se produjo ninguno"}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Errorf("Se esperaba error para tipo no soportado, pero no se produjo ninguno")
	} else {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Completado")
	}

	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestOrdenarJSON_JSON_Invalido(t *testing.T) {
	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, "cadena no válida")
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{TipoError: "JSON inválido"})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con JSON inválido")
	_, err := ordenJson.OrdenarJSON("cadena no válida")

	var actual ResultadosObtenidos
	if err == nil {
		actual = ResultadosObtenidos{Error: "Se esperaba error para JSON inválido, pero no se produjo ninguno"}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Errorf("Se esperaba error para JSON inválido, pero no se produjo ninguno")
	} else {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Completado")
	}

	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestOrdenarJSON_JSON_Vacio(t *testing.T) {
	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, "{}")
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: []string{}})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con JSON vacío")
	result, err := ordenJson.OrdenarJSON("{}")

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatalf("Error inesperado: %v", err)
	}

	trimmed := strings.TrimSpace(result)
	actual = ResultadosObtenidos{
		JsonSalida: trimmed,
	}

	status := "Completado"
	if trimmed != "{}" {
		status = "Fallido"
		t.Errorf("Se esperaba {} pero se obtuvo %s", trimmed)
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestOrdenarJSON_Campos_Extra(t *testing.T) {
	input := `{
		"extra:field1": "value1",
		"tanner:tipo-documento": "docType",
		"extra:field2": "value2",
		"tanner:rut-cliente": "12345678-9",
		"cm:description": "desc"
	}`

	expectedDefined := []string{"tanner:tipo-documento", "tanner:rut-cliente", "cm:description"}

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, input)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: expectedDefined})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con campos extra")
	result, err := ordenJson.OrdenarJSON(input)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatalf("Error inesperado: %v", err)
	}

	keys := extraerClavesJSON(result)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  result,
	}

	status := "Completado"
	if len(keys) != 5 {
		status = "Fallido"
		t.Fatalf("Se esperaban 5 llaves en total, pero se obtuvieron %d", len(keys))
	}

	for i, key := range expectedDefined {
		if keys[i] != key {
			status = "Fallido"
			t.Errorf("Se esperaba %s en la posición %d, pero se obtuvo %s", key, i, keys[i])
		}
	}

	extras := []string{keys[3], keys[4]}
	extraSet := map[string]bool{
		"extra:field1": true,
		"extra:field2": true,
	}
	for _, key := range extras {
		if !extraSet[key] {
			status = "Fallido"
			t.Errorf("Campo extra inesperado: %s", key)
		}
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestOrdenarJSON_Map_Extenso(t *testing.T) {
	inputMap := map[string]interface{}{
		"cm:description":       "desc",
		"tanner:estado-visado": "aprobado",
		"tanner:tipo-documento": "invoice",
		"cm:title":             "Invoice Title",
		"tanner:categorias":    "finance",
		"tanner:fecha-carga":   "2025-01-01T00:00:00.000Z",
		"tanner:nombre-doc":    "Document Name",
		"tanner:observaciones": "ninguna",
		"extra:1":              "val1",
		"extra:2":              "val2",
	}

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, inputMap)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{
		ClavesOrdenadas: []string{
			"tanner:tipo-documento",
			"tanner:estado-visado",
			"tanner:fecha-carga",
			"tanner:nombre-doc",
			"tanner:categorias",
			"cm:title",
			"cm:description",
			"tanner:observaciones",
		},
	})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con mapa extenso")
	result, err := ordenJson.OrdenarJSON(inputMap)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatalf("Error inesperado: %v", err)
	}

	keys := extraerClavesJSON(result)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  result,
	}

	var definedKeys []string
	var extraKeys []string
	for _, key := range keys {
		if obtenerOrden(key) < len(ordenJson.OrdenCampos) {
			definedKeys = append(definedKeys, key)
		} else {
			extraKeys = append(extraKeys, key)
		}
	}

	status := "Completado"
	if !reflect.DeepEqual(definedKeys, registradorGlobal.logs[testName].ResultadosEsperados.ClavesOrdenadas) {
		status = "Fallido"
		t.Errorf("Se esperaba el orden definido %v, pero se obtuvo %v", registradorGlobal.logs[testName].ResultadosEsperados.ClavesOrdenadas, definedKeys)
	}

	extraSet := map[string]bool{"extra:1": true, "extra:2": true}
	if len(extraKeys) != len(extraSet) {
		status = "Fallido"
		t.Errorf("Se esperaban %d llaves extras, pero se obtuvieron %d", len(extraSet), len(extraKeys))
	}

	for _, key := range extraKeys {
		if !extraSet[key] {
			status = "Fallido"
			t.Errorf("Llave extra inesperada: %s", key)
		}
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func obtenerOrden(campo string) int {
	for i, c := range ordenJson.OrdenCampos {
		if c == campo {
			return i
		}
	}
	return len(ordenJson.OrdenCampos)
}

func TestCaracteresEspecialesEnValores(t *testing.T) {
	input := `{
		"tanner:tipo-documento": "a\\b\"c\u00f1",
		"cm:description": "valor con \n salto de línea"
	}`

	expected := []string{"tanner:tipo-documento", "cm:description"}

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, input)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: expected})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con caracteres especiales")
	got, err := ordenJson.OrdenarJSON(input)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatal(err)
	}

	keys := extraerClavesJSON(got)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  got,
	}

	status := "Completado"
	if !reflect.DeepEqual(keys, expected) {
		status = "Fallido"
		t.Errorf("Claves esperadas: %v, obtenidas: %v", expected, keys)
	}

	if !strings.Contains(got, `"a\\b\"cñ"`) || !strings.Contains(got, `"valor con \n salto de línea"`) {
		status = "Fallido"
		t.Error("Caracteres especiales mal escapados")
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestJSONGrande(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("{")
	for i := 0; i < 100; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		key := fmt.Sprintf("campo%d", i)
		if i < 20 {
			key = ordenJson.OrdenCampos[i%len(ordenJson.OrdenCampos)]
		}
		sb.WriteString(fmt.Sprintf(`"%s": "valor%d"`, key, i))
	}
	sb.WriteString("}")

	input := sb.String()

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, input)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: ordenJson.OrdenCampos})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con JSON grande")
	got, err := ordenJson.OrdenarJSON(input)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatal(err)
	}

	keys := extraerClavesJSON(got)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  got,
	}

	status := "Completado"
	for i, key := range ordenJson.OrdenCampos {
		if i >= 20 {
			break
		}
		if keys[i] != key {
			status = "Fallido"
			t.Errorf("Posición %d: esperado %s, obtenido %s", i, key, keys[i])
		}
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestCamposNoDefinidos(t *testing.T) {
	input := `{
		"zzz": "debe ir al final",
		"tanner:rut-cliente": "123",
		"aaa": "debe ir después de los definidos"
	}`

	expectedOrder := []string{
		"tanner:rut-cliente",
		"zzz",
		"aaa",
	}

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, input)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: expectedOrder})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con campos no definidos")
	got, err := ordenJson.OrdenarJSON(input)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatal(err)
	}

	keys := extraerClavesJSON(got)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  got,
	}

	status := "Completado"
	if !reflect.DeepEqual(keys, expectedOrder) {
		status = "Fallido"
		t.Errorf("Orden incorrecto. Esperado: %v, Obtenido: %v", expectedOrder, keys)
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestTiposDeDatosVariados(t *testing.T) {
	input := `{
		"tanner:tipo-documento": 123,
		"cm:title": true,
		"tanner:origen": null,
		"cm:versionLabel": [1, "dos", false]
	}`

	expectedKeys := []string{"tanner:tipo-documento", "tanner:origen", "cm:title", "cm:versionLabel"}

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, input)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: expectedKeys})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con tipos de datos variados")
	got, err := ordenJson.OrdenarJSON(input)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatal(err)
	}

	keys := extraerClavesJSON(got)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  got,
	}

	status := "Completado"
	if !reflect.DeepEqual(keys, expectedKeys) {
		status = "Fallido"
		t.Errorf("Orden de claves incorrecto: %v", keys)
	}

	if !strings.Contains(got, `123`) || !strings.Contains(got, `true`) || !strings.Contains(got, `null`) {
		status = "Fallido"
		t.Error("Valores no serializados correctamente")
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func TestJSONMalformado(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "Falta llave de cierre", input: `{"tanner:tipo-documento": "test"`},
		{name: "Clave sin comillas", input: `{tanner:tipo-documento: "test"}`},
		{name: "Valor sin cerrar", input: `{"cm:title": "test`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testName := t.Name()
			startTime := time.Now()
			registradorGlobal.IniciadorTest(testName, tt.input)
			registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{TipoError: "JSON malformado"})

			registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarJSON con JSON malformado")
			_, err := ordenJson.OrdenarJSON(tt.input)

			var actual ResultadosObtenidos
			if err == nil {
				actual = ResultadosObtenidos{Error: "Se esperaba un error por JSON malformado"}
				registradorGlobal.GuardarResultado(testName, actual, "Fallido")
				t.Error("Se esperaba un error por JSON malformado")
			} else {
				actual = ResultadosObtenidos{Error: err.Error()}
				registradorGlobal.GuardarResultado(testName, actual, "Completado")
			}

			registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
		})
	}
}

func TestJSONCamposVacios(t *testing.T) {
	metadata := ordenJson.DocumentMetadata{
		TipoDocumento: "",
		RUTCliente:   "123",
		CmTitle:      "",
		Origen:       "central",
	}

	expectedOrder := []string{"tanner:rut-cliente", "tanner:origen"}

	testName := t.Name()
	startTime := time.Now()
	registradorGlobal.IniciadorTest(testName, metadata)
	registradorGlobal.ConfigResultadoEsperado(testName, ResultadosEsperados{ClavesOrdenadas: expectedOrder})

	registradorGlobal.AgregarProceso(testName, "Ejecutando OrdenarDocumentoMetadata con campos vacíos")
	got, err := ordenJson.OrdenarDocumentoMetadata(metadata)

	var actual ResultadosObtenidos
	if err != nil {
		actual = ResultadosObtenidos{Error: err.Error()}
		registradorGlobal.GuardarResultado(testName, actual, "Fallido")
		t.Fatal(err)
	}

	keys := extraerClavesJSON(got)
	actual = ResultadosObtenidos{
		ClavesOrdenadas: keys,
		JsonSalida:  got,
	}

	status := "Completado"
	if !reflect.DeepEqual(keys, expectedOrder) {
		status = "Fallido"
		t.Errorf("Campos vacíos no filtrados. Claves: %v", keys)
	}

	registradorGlobal.GuardarResultado(testName, actual, status)
	registradorGlobal.logs[testName].TiempoDeEjecucion = time.Since(startTime).String()
}

func BenchmarkOrdenarJSON(b *testing.B) {
	input := `{"zzz": "valor", "tanner:tipo-documento": "test", "cm:title": "title"}`

	for i := 0; i < b.N; i++ {
		_, _ = ordenJson.OrdenarJSON(input)
	}
}


// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// ~ HOOK PARA GUARDAR LOS LOGS AL FINAL ~
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestMain(m *testing.M) {
	code := m.Run()
	if err := registradorGlobal.CrearArchivoLog(); err != nil {
		fmt.Printf("Error escribiendo logs: %v\n", err)
	}
	os.Exit(code)
}
