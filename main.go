package main

import (
	"fmt"

	"github.com/samuel/prueba-orden/ordenJson"
)

func main() {
	// JSON desordenado con todos los campos del struct DocumentMetadata
	jsonDesordenado := `{
		"cm:description": "Descripción del documento",
		"tanner:rut-cliente": "12345678-9",
		"tanner:tipo-documento": "contrato",
		"cm:title": "Título del documento",
		"tanner:origen": "Departamento Legal",
		"tanner:razon-social-cliente": "Empresa Ejemplo S.A.",
		"tanner:estado-visado": "aprobado",
		"tanner:estado-vigencia": "vigente",
		"tanner:fecha-carga": "2023-10-01T00:00:00.000Z",
		"tanner:nombre-doc": "Documento de Ejemplo",
		"tanner:categorias": "legal",
		"tanner:sub-categorias": "contratos",
		"tanner:relacion": "cliente",
		"tanner:fecha-termino-vigencia": "2024-10-01T00:00:00.000Z",
		"cm:versionType": "1.0",
		"cm:versionLabel": "v1.0",
		"tanner:observaciones": "Ninguna"
	}`

	// Ordenar el JSON utilizando la función OrdenarJSON del paquete ordenJson
	jsonOrdenado, err := ordenJson.OrdenarJSON(jsonDesordenado)
	if err != nil {
		fmt.Printf("Error al ordenar el JSON: %v\n", err)
		return
	}

	// Imprimir el JSON ordenado en la terminal
	fmt.Println("JSON ordenado:")
	fmt.Println(jsonOrdenado)
}
