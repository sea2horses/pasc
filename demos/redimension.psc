// Redimensionar conserva los elementos que siguen cabiendo
Algoritmo Redimension
	Definir i, n Como Entero
	n <- 3
	Dimensionar lista[n]
	Para i <- 0 Hasta n - 1 Hacer
		lista[i] <- (i + 1) * 10
	FinPara

	Redimensionar lista[5]
	lista[3] <- 40
	lista[4] <- 50

	Para i <- 0 Hasta 4 Hacer
		Escribir Sin Saltar lista[i], " "
	FinPara
	Escribir ""
FinAlgoritmo
