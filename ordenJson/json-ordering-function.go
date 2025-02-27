package ordenJson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"reflect"
)

// DocumentMetadata representa la estructura de metadatos del documento.
// Cada campo tiene etiquetas JSON y BSON para facilitar la serialización y deserialización.
type DocumentMetadata struct {
	TipoDocumento        string `json:"tanner:tipo-documento" bson:"_tanner:tipo-documento, omitempty"` // Tipo de documento (ej: contrato, factura)
	RazonSocialCliente   string `json:"tanner:razon-social-cliente" bson:"_tanner:razon-social-cliente, omitempty"` // Razón social del cliente
	RUTCliente           string `json:"tanner:rut-cliente" bson:"_tanner:rut-cliente, omitempty"` // RUT del cliente
	EstadoVisado         string `json:"tanner:estado-visado" bson:"_tanner:estado-visado, omitempty"` // Estado de visado (ej: aprobado, rechazado)
	EstadoVigencia       string `json:"tanner:estado-vigencia" bson:"_tanner:estado-vigencia, omitempty"` // Estado de vigencia (ej: vigente, vencido)
	FechaCarga           string `json:"tanner:fecha-carga" bson:"_tanner:fecha-carga, omitempty" validate:"datetime=2006-01-02T15:04:05.999Z07:00"` // Fecha de carga del documento
	NombreDoc            string `json:"tanner:nombre-doc" bson:"_tanner:nombre-doc, omitempty"` // Nombre del documento
	Categorias           string `json:"tanner:categorias" bson:"_tanner:categorias, omitempty"` // Categoría del documento
	SubCategorias        string `json:"tanner:sub-categorias" bson:"_tanner:sub-categorias, omitempty"` // Subcategoría del documento
	Origen               string `json:"tanner:origen" bson:"_tanner:origen, omitempty"` // Origen del documento (ej: departamento legal)
	Relacion             string `json:"tanner:relacion" bson:"_tanner:relacion, omitempty"` // Relación del documento (ej: cliente, proveedor)
	FechaTerminoVigencia string `json:"tanner:fecha-termino-vigencia" bson:"_tanner:fecha-termino-vigencia, omitempty"` // Fecha de término de vigencia
	CmTitle              string `json:"cm:title" bson:"_cm:title, omitempty"` // Título del documento
	CmVersionType        string `json:"cm:versionType" bson:"_cm:versionType, omitempty"` // Tipo de versión del documento
	CmVersionLabel       string `json:"cm:versionLabel" bson:"_cm:versionLabel, omitempty"` // Etiqueta de versión del documento
	CmDescription        string `json:"cm:description" bson:"_cm:description, omitempty"` // Descripción del documento
	Observaciones        string `json:"tanner:observaciones" bson:"_tanner:observaciones,omitempty"` // Observaciones adicionales
}

// OrdenCampos define el orden deseado de los campos en el JSON.
// El índice en el slice representa la prioridad (menor índice = mayor prioridad).
var OrdenCampos = []string{
	"tanner:tipo-documento",
	"tanner:razon-social-cliente",
	"tanner:rut-cliente",
	"tanner:estado-visado",
	"tanner:estado-vigencia",
	"tanner:fecha-carga",
	"tanner:nombre-doc",
	"tanner:categorias",
	"tanner:sub-categorias",
	"tanner:origen",
	"tanner:relacion",
	"tanner:fecha-termino-vigencia",
	"cm:title",
	"cm:versionType",
	"cm:versionLabel",
	"cm:description",
	"tanner:observaciones",
}

// ordenCampoMap es un mapa que almacena la posición de cada campo en OrdenCampos.
// Se utiliza para optimizar la búsqueda de la posición de un campo durante la ordenación.
var ordenCampoMap map[string]int

// init inicializa el mapa ordenCampoMap con las posiciones de los campos en OrdenCampos.
// Esto permite una búsqueda rápida de la posición de un campo durante la ordenación.
func init() {
	ordenCampoMap = make(map[string]int, len(OrdenCampos))
	for i, campo := range OrdenCampos {
		ordenCampoMap[campo] = i
	}
}

// obtenerOrdenCampo devuelve la posición de un campo usando el mapa precalculado.
// Si el campo no está en la lista, retorna la longitud de la lista, ubicándolo al final.
func obtenerOrdenCampo(campo string) int {
	if orden, ok := ordenCampoMap[campo]; ok {
		return orden
	}
	return len(OrdenCampos)
}

// OrdenarDocumentoMetadata recibe un DocumentMetadata y devuelve un JSON ordenado.
// Filtra los campos vacíos y ordena los campos según el orden predefinido.
func OrdenarDocumentoMetadata(metadata DocumentMetadata) (string, error) {
	// Crear un mapa para incluir solo los campos no vacíos.
	datos := make(map[string]interface{})

	// Usar reflexión para iterar sobre los campos del struct.
	val := reflect.ValueOf(metadata)
	typ := reflect.TypeOf(metadata)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)       // Valor del campo
		fieldType := typ.Field(i)  // Tipo del campo (incluye etiquetas)

		// Obtener la etiqueta JSON del campo.
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			continue // Si no tiene etiqueta JSON, se ignora.
		}

		// Verificar si el campo no está vacío.
		if field.String() != "" {
			datos[jsonTag] = field.String() // Agregar al mapa si no está vacío.
		}
	}

	// Ordenar el JSON utilizando la función OrdenarJSON.
	return OrdenarJSON(datos)
}
// OrdenarJSON recibe un JSON desordenado (como cadena o mapa) y lo devuelve ordenado según el orden predefinido.
// Si el input es una cadena, se convierte a un mapa antes de ordenar.
func OrdenarJSON(input interface{}) (string, error) {
	var datos map[string]interface{}

	// Convertir el input a un mapa.
	switch v := input.(type) {
	case string:
		// Si el input es una cadena, convertirla a un mapa.
		if err := json.Unmarshal([]byte(v), &datos); err != nil {
			return "", err
		}
	case map[string]interface{}:
		// Si el input ya es un mapa, usarlo directamente.
		datos = v
	default:
		// Si el tipo de entrada no es soportado, retornar un error.
		return "", fmt.Errorf("tipo de entrada no soportado: %T", input)
	}

	// Obtener las claves del mapa.
	claves := make([]string, 0, len(datos))
	for clave := range datos {
		claves = append(claves, clave)
	}

	// Ordenar las claves según el orden predefinido.
	sort.Slice(claves, func(i, j int) bool {
		return obtenerOrdenCampo(claves[i]) < obtenerOrdenCampo(claves[j])
	})

	// Construir manualmente el JSON ordenado usando bytes.Buffer.
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, clave := range claves {
		if i > 0 {
			buf.WriteByte(',')
		}
		// Codificar la clave.
		claveJSON, err := json.Marshal(clave)
		if err != nil {
			return "", err
		}
		buf.Write(claveJSON)
		buf.WriteByte(':')
		// Codificar el valor.
		valorJSON, err := json.Marshal(datos[clave])
		if err != nil {
			return "", err
		}
		buf.Write(valorJSON)
	}
	buf.WriteByte('}')

	// Formatear el JSON con indentación.
	var resultado bytes.Buffer
	if err := json.Indent(&resultado, buf.Bytes(), "", "  "); err != nil {
		return "", err
	}
	return resultado.String(), nil
}

// OrdenarMapaComoDocumentoMetadata convierte un mapa a JSON y luego lo ordena.
// Es un wrapper alrededor de OrdenarJSON para facilitar su uso con mapas.
func OrdenarMapaComoDocumentoMetadata(mapa map[string]interface{}) (string, error) {
	return OrdenarJSON(mapa)
}