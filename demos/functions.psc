Funcion suma <- Sumar(a, b)
	suma = a + b
Fin Funcion

Funcion es_mayor <- EsMayor(a, b)
	es_mayor = a > b
Fin Funcion

SubProceso Intercambiar(a Por Referencia, b Por Referencia)
	auxiliar = a
	a = b
	b = auxiliar
Fin SubProceso

Algoritmo PruebaFunciones
	Definir total Como Entero
	Definir mayor Como Logico
	Definir primero, segundo Como Entero

	total = Sumar(12, 30)
	mayor = EsMayor(total, 20)
	primero = 3
	segundo = 8
	Intercambiar(primero, segundo)

	Escribir "La suma es: ", total
	Escribir "¿La suma supera 20? ", mayor
	Escribir "Valores intercambiados: ", primero, ", ", segundo
FinAlgoritmo
