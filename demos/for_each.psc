// Para Cada recorre todos los elementos; la variable es una referencia
Algoritmo RecorrerArreglos
	Definir i, total, precios Como Entero
	Dimensionar precios[4]
	Para i <- 0 Hasta 3 Hacer
		precios[i] <- (i + 1) * 100
	FinPara

	total <- 0
	Para Cada precio De precios Hacer
		total <- total + precio
		precio <- precio * 2
	FinPara
	Escribir "Total original: ", total

	Escribir Sin Saltar "Precios duplicados: "
	Para Cada precio De precios Hacer
		Escribir Sin Saltar precio, " "
	FinPara
	Escribir ""
FinAlgoritmo
