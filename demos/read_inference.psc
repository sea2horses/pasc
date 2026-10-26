// `Leer` can introduce variables without a prior `Definir` statement.
// `sin_tipo` falls back to cadena, `cantidad` is inferred as entero, and
// `mensaje` is inferred as cadena from string concatenation.
Algoritmo InferirTiposConLeer
	Escribir "Escribe tres valores: uno libre, un numero y un mensaje."
	Leer sin_tipo, cantidad, mensaje

	cantidad = 42
	mensaje = mensaje + "!"

	Escribir "Valor sin contexto: ", sin_tipo
	Escribir "Cantidad inferida: ", cantidad
	Escribir "Mensaje inferido: ", mensaje
FinAlgoritmo
