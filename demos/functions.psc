Funcion suma <- Sumar(a, b)
	suma = a + b
Fin Funcion

Funcion es_mayor <- EsMayor(a, b)
	es_mayor = a > b
Fin Funcion

Algoritmo PruebaFunciones
	Definir total Como Entero
	Definir mayor Como Logico

	total = Sumar(12, 30)
	mayor = EsMayor(total, 20)

	Escribir "La suma es: ", total
	Escribir "¿La suma supera 20? ", mayor
FinAlgoritmo
