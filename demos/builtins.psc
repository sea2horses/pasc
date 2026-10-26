Algoritmo PruebaFuncionesIntegradas
	Definir texto Como Cadena
	Definir numero Como Real

	texto = Concatenar("Hola, ", "PSeInt")
	Escribir Mayusculas(texto)
	Escribir Minusculas(texto)
	Escribir "Caracteres: ", Longitud(texto)
	Escribir "Subcadena: ", Subcadena(texto, 0, 3)

	numero = ConvertirANumero("12.5")
	Escribir "Texto convertido a numero: ", numero
	Escribir "Numero convertido a texto: ", ConvertirATexto(numero)

	Escribir "PI: ", PI
	Escribir "RC de PI: ", RC(PI)
	Escribir "Raiz de PI: ", Raiz(PI)
	Escribir "Valor absoluto: ", Abs(-4.5)
	Escribir "Logaritmo natural: ", Ln(Exp(1))
	Escribir "Seno: ", Sen(PI / 2)
	Escribir "Coseno: ", Cos(0)
	Escribir "Tangente: ", Tan(0)
	Escribir "Arcoseno: ", ASen(1)
	Escribir "Arcocoseno: ", ACos(1)
	Escribir "Arcotangente: ", ATan(0)
	Escribir "Truncamiento: ", Trunc(3.8)
	Escribir "Redondeo: ", Redon(3.8)
	Escribir "Azar entre 0 y 5: ", Azar(6)
	Escribir "Aleatorio entre 10 y 20: ", Aleatorio(10, 20)
	Escribir "Hora: ", HoraActual()
	Escribir "Fecha: ", FechaActual()
FinAlgoritmo
