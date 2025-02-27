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

// Evento define la estructura de un registro de evento.
type Evento struct {
	Timestamp string                 `json:"timestamp"`   // Fecha y hora en RFC3339
	TestName  string                 `json:"testName"`    // Nombre del test
	EventType string                 `json:"eventType"`   // INFO, DEBUG, ERROR
	Details   map[string]interface{} `json:"details"`     // Datos adicionales
}

// TestLogger centraliza el registro de eventos durante las pruebas.
type TestLogger struct {
	mu      sync.Mutex
	eventos []Evento
}

var globalLogger = &TestLogger{}

// Log registra un evento de manera segura para concurrencia.
func (tl *TestLogger) Log(testName, eventType string, details map[string]interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	tl.eventos = append(tl.eventos, Evento{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		TestName:  testName,
		EventType: eventType,
		Details:   details,
	})
}

// WriteLogsToFile escribe todos los eventos en un archivo JSON.
func (tl *TestLogger) WriteLogsToFile() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	file, err := json.MarshalIndent(tl.eventos, "", "  ")
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

func extractKeys(orderedJSON string) []string {
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

			// Registro: Inicio del test
			globalLogger.Log(testName, "INFO", map[string]interface{}{
				"accion":   "Inicio del test",
				"input":    tt.input,
				"expected": tt.expected,
			})

			// Registro: Ejecución de la función
			globalLogger.Log(testName, "INFO", map[string]interface{}{
				"accion": "Ejecutando OrdenarJSON",
			})

			got, err := ordenJson.OrdenarJSON(tt.input)

			if err != nil {
				// Registro: Error en la función
				globalLogger.Log(testName, "ERROR", map[string]interface{}{
					"accion": "OrdenarJSON falló",
					"error":  err.Error(),
				})
				t.Fatalf("OrdenarJSON() error = %v", err)
			}

			// Registro: Resultado obtenido
			globalLogger.Log(testName, "DEBUG", map[string]interface{}{
				"accion": "Resultado generado",
				"output": got,
			})

			keys := extractKeys(got)
			if !reflect.DeepEqual(keys, tt.expected) {
				// Registro: Error de aserción
				globalLogger.Log(testName, "ERROR", map[string]interface{}{
					"accion":   "Comparación de claves fallida",
					"esperado": tt.expected,
					"obtenido": keys,
				})
				t.Errorf("Orden de claves incorrecto")
			}

			// Registro: Conclusión del test
			globalLogger.Log(testName, "INFO", map[string]interface{}{
				"accion": "Test finalizado",
				"estado": "Éxito",
			})
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

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion":   "Inicio del test",
		"metadata": metadata,
		"expected": expectedOrder,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarDocumentoMetadata",
	})

	orderedJSON, err := ordenJson.OrdenarDocumentoMetadata(metadata)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarDocumentoMetadata falló",
			"error":  err.Error(),
		})
		t.Fatalf("OrdenarDocumentoMetadata() error = %v", err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": orderedJSON,
	})

	keys := extractKeys(orderedJSON)
	if !reflect.DeepEqual(keys, expectedOrder) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Comparación de claves fallida",
			"esperado": expectedOrder,
			"obtenido": keys,
		})
		t.Errorf("Orden de claves incorrecto")
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
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

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion":   "Inicio del test",
		"inputMap": inputMap,
		"expected": expectedOrder,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarMapaComoDocumentoMetadata",
	})

	orderedJSON, err := ordenJson.OrdenarMapaComoDocumentoMetadata(inputMap)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarMapaComoDocumentoMetadata falló",
			"error":  err.Error(),
		})
		t.Fatalf("OrdenarMapaComoDocumentoMetadata() error = %v", err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": orderedJSON,
	})

	keys := extractKeys(orderedJSON)
	if !reflect.DeepEqual(keys, expectedOrder) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Comparación de claves fallida",
			"esperado": expectedOrder,
			"obtenido": keys,
		})
		t.Errorf("Orden de claves incorrecto")
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
}

func TestOrdenarJSON_UnsupportedType(t *testing.T) {
	testName := t.Name()

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Inicio del test",
		"input":  123,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con tipo no soportado",
	})

	_, err := ordenJson.OrdenarJSON(123)
	if err == nil {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "Se esperaba error para tipo no soportado, pero no se produjo ninguno",
		})
		t.Errorf("Se esperaba error para tipo no soportado, pero no se produjo ninguno")
	} else {
		// Registro: Error capturado
		globalLogger.Log(testName, "DEBUG", map[string]interface{}{
			"accion": "Error capturado",
			"error":  err.Error(),
		})
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
}

func TestOrdenarJSON_InvalidJSONString(t *testing.T) {
	testName := t.Name()

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Inicio del test",
		"input":  "cadena no válida",
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con JSON inválido",
	})

	_, err := ordenJson.OrdenarJSON("cadena no válida")
	if err == nil {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "Se esperaba error para JSON inválido, pero no se produjo ninguno",
		})
		t.Errorf("Se esperaba error para JSON inválido, pero no se produjo ninguno")
	} else {
		// Registro: Error capturado
		globalLogger.Log(testName, "DEBUG", map[string]interface{}{
			"accion": "Error capturado",
			"error":  err.Error(),
		})
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
}

func TestOrdenarJSON_EmptyJSON(t *testing.T) {
	testName := t.Name()

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Inicio del test",
		"input":  "{}",
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con JSON vacío",
	})

	result, err := ordenJson.OrdenarJSON("{}")
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarJSON falló",
			"error":  err.Error(),
		})
		t.Fatalf("Error inesperado: %v", err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": result,
	})

	trimmed := strings.TrimSpace(result)
	if trimmed != "{}" {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Comparación de JSON fallida",
			"esperado": "{}",
			"obtenido": trimmed,
		})
		t.Errorf("Se esperaba {} pero se obtuvo %s", trimmed)
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
}

func TestOrdenarJSON_ExtraFields(t *testing.T) {
	input := `{
		"extra:field1": "value1",
		"tanner:tipo-documento": "docType",
		"extra:field2": "value2",
		"tanner:rut-cliente": "12345678-9",
		"cm:description": "desc"
	}`

	expectedDefined := []string{"tanner:tipo-documento", "tanner:rut-cliente", "cm:description"}

	testName := t.Name()

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Inicio del test",
		"input":  input,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con campos extra",
	})

	result, err := ordenJson.OrdenarJSON(input)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarJSON falló",
			"error":  err.Error(),
		})
		t.Fatalf("Error inesperado: %v", err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": result,
	})

	keys := extractKeys(result)
	if len(keys) != 5 {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Número incorrecto de claves",
			"esperado": 5,
			"obtenido": len(keys),
		})
		t.Fatalf("Se esperaban 5 llaves en total, pero se obtuvieron %d", len(keys))
	}

	for i, key := range expectedDefined {
		if keys[i] != key {
			// Registro: Error de aserción
			globalLogger.Log(testName, "ERROR", map[string]interface{}{
				"accion":   "Clave definida en posición incorrecta",
				"esperado": key,
				"obtenido": keys[i],
				"posicion": i,
			})
			t.Errorf("Se esperaba %s en la posición %d, pero se obtuvo %s", key, i, keys[i])
		}
	}

	// Verificar campos extras
	extras := []string{keys[3], keys[4]}
	extraSet := map[string]bool{
		"extra:field1": true,
		"extra:field2": true,
	}
	for _, key := range extras {
		if !extraSet[key] {
			// Registro: Error de aserción
			globalLogger.Log(testName, "ERROR", map[string]interface{}{
				"accion":   "Campo extra inesperado",
				"campo":    key,
			})
			t.Errorf("Campo extra inesperado: %s", key)
		}
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
}

func TestOrdenarJSON_MapExtensive(t *testing.T) {
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

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion":   "Inicio del test",
		"inputMap": inputMap,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con mapa extenso",
	})

	result, err := ordenJson.OrdenarJSON(inputMap)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarJSON falló",
			"error":  err.Error(),
		})
		t.Fatalf("Error inesperado: %v", err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": result,
	})

	keys := extractKeys(result)

	var definedKeys []string
	var extraKeys []string
	for _, key := range keys {
		if getOrder(key) < len(ordenJson.OrdenCampos) {
			definedKeys = append(definedKeys, key)
		} else {
			extraKeys = append(extraKeys, key)
		}
	}

	expectedDefined := []string{
		"tanner:tipo-documento",
		"tanner:estado-visado",
		"tanner:fecha-carga",
		"tanner:nombre-doc",
		"tanner:categorias",
		"cm:title",
		"cm:description",
		"tanner:observaciones",
	}

	if !reflect.DeepEqual(definedKeys, expectedDefined) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Orden de claves definidas incorrecto",
			"esperado": expectedDefined,
			"obtenido": definedKeys,
		})
		t.Errorf("Se esperaba el orden definido %v, pero se obtuvo %v", expectedDefined, definedKeys)
	}

	extraSet := map[string]bool{"extra:1": true, "extra:2": true}
	if len(extraKeys) != len(extraSet) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Número incorrecto de claves extras",
			"esperado": len(extraSet),
			"obtenido": len(extraKeys),
		})
		t.Errorf("Se esperaban %d llaves extras, pero se obtuvieron %d", len(extraSet), len(extraKeys))
	}

	for _, key := range extraKeys {
		if !extraSet[key] {
			// Registro: Error de aserción
			globalLogger.Log(testName, "ERROR", map[string]interface{}{
				"accion":   "Llave extra inesperada",
				"llave":    key,
			})
			t.Errorf("Llave extra inesperada: %s", key)
		}
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
}

func getOrder(campo string) int {
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

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Inicio del test",
		"input":  input,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con caracteres especiales",
	})

	got, err := ordenJson.OrdenarJSON(input)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarJSON falló",
			"error":  err.Error(),
		})
		t.Fatal(err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": got,
	})

	keys := extractKeys(got)
	if !reflect.DeepEqual(keys, expected) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Comparación de claves fallida",
			"esperado": expected,
			"obtenido": keys,
		})
		t.Errorf("Claves esperadas: %v, obtenidas: %v", expected, keys)
	}

	// Verificar que los valores no se corrompan
	if !strings.Contains(got, `"a\\b\"cñ"`) || !strings.Contains(got, `"valor con \n salto de línea"`) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Caracteres especiales mal escapados",
			"output":   got,
		})
		t.Error("Caracteres especiales mal escapados")
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
}

func TestJSONGrande(t *testing.T) {
	// Generar un JSON con 100 campos (20 definidos + 80 aleatorios)
	var sb strings.Builder
	sb.WriteString("{")
	for i := 0; i < 100; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		key := fmt.Sprintf("campo%d", i)
		if i < 20 { // Los primeros 20 están en OrdenCampos
			key = ordenJson.OrdenCampos[i%len(ordenJson.OrdenCampos)]
		}
		sb.WriteString(fmt.Sprintf(`"%s": "valor%d"`, key, i))
	}
	sb.WriteString("}")

	input := sb.String()

	testName := t.Name()

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Inicio del test",
		"input":  input,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con JSON grande",
	})

	got, err := ordenJson.OrdenarJSON(input)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarJSON falló",
			"error":  err.Error(),
		})
		t.Fatal(err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": got,
	})

	keys := extractKeys(got)
	// Verificar que los primeros 20 campos están ordenados según OrdenCampos
	for i, key := range ordenJson.OrdenCampos {
		if i >= 20 {
			break
		}
		if keys[i] != key {
			// Registro: Error de aserción
			globalLogger.Log(testName, "ERROR", map[string]interface{}{
				"accion":   "Clave en posición incorrecta",
				"esperado": key,
				"obtenido": keys[i],
				"posicion": i,
			})
			t.Errorf("Posición %d: esperado %s, obtenido %s", i, key, keys[i])
		}
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
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
		"aaa", // Los no definidos mantienen su orden relativo
	}

	testName := t.Name()

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Inicio del test",
		"input":  input,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con campos no definidos",
	})

	got, err := ordenJson.OrdenarJSON(input)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarJSON falló",
			"error":  err.Error(),
		})
		t.Fatal(err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": got,
	})

	keys := extractKeys(got)
	if !reflect.DeepEqual(keys, expectedOrder) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Orden incorrecto",
			"esperado": expectedOrder,
			"obtenido": keys,
		})
		t.Errorf("Orden incorrecto. Esperado: %v, Obtenido: %v", expectedOrder, keys)
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
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

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Inicio del test",
		"input":  input,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarJSON con tipos de datos variados",
	})

	got, err := ordenJson.OrdenarJSON(input)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarJSON falló",
			"error":  err.Error(),
		})
		t.Fatal(err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": got,
	})

	keys := extractKeys(got)
	if !reflect.DeepEqual(keys, expectedKeys) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Orden de claves incorrecto",
			"esperado": expectedKeys,
			"obtenido": keys,
		})
		t.Errorf("Orden de claves incorrecto: %v", keys)
	}

	// Validar tipos
	if !strings.Contains(got, `123`) || !strings.Contains(got, `true`) || !strings.Contains(got, `null`) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Valores no serializados correctamente",
			"output":   got,
		})
		t.Error("Valores no serializados correctamente")
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
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

			// Registro: Inicio del test
			globalLogger.Log(testName, "INFO", map[string]interface{}{
				"accion": "Inicio del test",
				"input":  tt.input,
			})

			// Registro: Ejecución de la función
			globalLogger.Log(testName, "INFO", map[string]interface{}{
				"accion": "Ejecutando OrdenarJSON con JSON malformado",
			})

			_, err := ordenJson.OrdenarJSON(tt.input)
			if err == nil {
				// Registro: Error de aserción
				globalLogger.Log(testName, "ERROR", map[string]interface{}{
					"accion": "Se esperaba un error por JSON malformado",
				})
				t.Error("Se esperaba un error por JSON malformado")
			} else {
				// Registro: Error capturado
				globalLogger.Log(testName, "DEBUG", map[string]interface{}{
					"accion": "Error capturado",
					"error":  err.Error(),
				})
			}

			// Registro: Conclusión del test
			globalLogger.Log(testName, "INFO", map[string]interface{}{
				"accion": "Test finalizado",
				"estado": "Éxito",
			})
		})
	}
}

func TestCamposVacios(t *testing.T) {
	metadata := ordenJson.DocumentMetadata{
		TipoDocumento: "", // Vacío (no debe aparecer)
		RUTCliente:   "123",
		CmTitle:      "", // Vacío (no debe aparecer)
		Origen:       "central",
	}

	expectedOrder := []string{"tanner:rut-cliente", "tanner:origen"}

	testName := t.Name()

	// Registro: Inicio del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion":   "Inicio del test",
		"metadata": metadata,
	})

	// Registro: Ejecución de la función
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Ejecutando OrdenarDocumentoMetadata con campos vacíos",
	})

	got, err := ordenJson.OrdenarDocumentoMetadata(metadata)
	if err != nil {
		// Registro: Error en la función
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion": "OrdenarDocumentoMetadata falló",
			"error":  err.Error(),
		})
		t.Fatal(err)
	}

	// Registro: Resultado obtenido
	globalLogger.Log(testName, "DEBUG", map[string]interface{}{
		"accion": "Resultado generado",
		"output": got,
	})

	keys := extractKeys(got)
	if !reflect.DeepEqual(keys, expectedOrder) {
		// Registro: Error de aserción
		globalLogger.Log(testName, "ERROR", map[string]interface{}{
			"accion":   "Campos vacíos no filtrados",
			"esperado": expectedOrder,
			"obtenido": keys,
		})
		t.Errorf("Campos vacíos no filtrados. Claves: %v", keys)
	}

	// Registro: Conclusión del test
	globalLogger.Log(testName, "INFO", map[string]interface{}{
		"accion": "Test finalizado",
		"estado": "Éxito",
	})
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
	if err := globalLogger.WriteLogsToFile(); err != nil {
		fmt.Printf("Error escribiendo logs: %v\n", err)
	}
	os.Exit(code)
}