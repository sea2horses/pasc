Funcion valor <- PrimerElemento(valores)
	valor = valores[0]
Fin Funcion

Algoritmo RegresionesCompilador
	Definir valor Como Entero
	Dimension valores(1), sinTipo(1)
	valores[0] = 7
	valor = PrimerElemento(valores)
	Escribir "valor=", valor, ","; Sin Saltar
	Escribir " fin"
	Escribir, sinTipo[0]
FinAlgoritmo
