<div align="center">

![pasc-inverted-banner](assets/pasc-inverted-banner.svg)

# `PASC`

**Compilador experimental de pseudocódigo PSeInt**

[![Go](https://img.shields.io/badge/go-1.25%2B-blue?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Pseint](https://img.shields.io/badge/compilador%20de-pseint-yellow?style=for-the-badge&logoColor=white)](https://pseint.sourceforge.net/)
[![Plataforma](https://img.shields.io/badge/multiplataforma-3c3c3c?style=for-the-badge)](#requisitos)

</div>

`pascc` es un compilador y analizador experimental para el lenguaje de PSeInt. Analiza un archivo `.psc`, genera código Go y utiliza la herramienta de Go para crear un ejecutable nativo. La interfaz de línea de comandos está organizada con Cobra y ofrece dos comandos: `run` para compilar y ejecutar temporalmente, y `build` para conservar el ejecutable.

> [!IMPORTANT]
> El proyecto sigue en desarrollo. Ya puede ejecutar programas útiles, pero todavía no implementa todo el lenguaje de PSeInt ni debe considerarse compatible al 100 %.

## Ventajas

<div align="center">

- **Compilación nativa:** El pseudocódigo se transforma en Go y después en un ejecutable nativo.
![Compilation](assets/compilation.png)
- **Errores fáciles de ubicar:** Los diagnósticos muestran el archivo, la línea, la columna, el fragmento afectado y un resumen de errores y advertencias.
![Diagnostic2](assets/diagnostic-2.png)
- **Análisis a tiempo de compilación:** Valida operaciones y asignaciones antes de generar el programa, con soporte para conversiones numéricas seguras.
![Diagnostic1](assets/diagnostic-1.png)
![Diagnostic3](assets/diagnostic-3.png)
- **Sintaxis familiar:** Implementa la sintaxis de PSeInt.
- **Rápidez:** El código generado se compila a un ejecutable nativo.<br>
![DonutDemo](assets/donut-demo.gif)

La dona animada es un programa de PSeInt compilado con `pascc`: [ver demo](demos/donut.psc).
- **Arquitectura modular:** El analizador, el generador, el runtime y los comandos están organizados en paquetes separados.

</div>

## Funcionalidades disponibles

Actualmente se soportan:

- Programas delimitados por `Algoritmo` y `FinAlgoritmo`.
- Tipos `Entero`, `Real`, `Cadena` y `Logico`.
- Declaraciones explícitas y asignaciones.
- Literales numéricos, cadenas y valores lógicos.
- `Escribir`, `Leer` y `Borrar Pantalla`.
- Condicionales `Si`, `SiNo` y `FinSi`.
- Ciclos `Mientras ... Hacer` y `FinMientras`.
- Selección con `Segun`, casos, `De Otro Modo` y `FinSegun`.
- Operadores aritméticos, relacionales y lógicos, incluyendo módulo.
- Paréntesis y precedencia de operadores.

## Requisitos

- [Go 1.25 o posterior](https://go.dev/dl/).
- Git, únicamente si vas a clonar el repositorio.
- [Just](https://just.systems/) es opcional y facilita algunos comandos de desarrollo.

Para compilar programas `.psc`, `pascc` ejecuta `go build`, así que Go debe estar disponible en el `PATH`. Los ejecutables creados con `build` se ejecutan sin necesitar Go.

## Instalación

Clona el repositorio y compila la herramienta:

```bash
git clone https://github.com/sea2horses/pasc.git
cd pasc
go build -o pascc .
```

En Windows, el ejecutable se llamará `pascc.exe`. También puedes instalarlo en el directorio de binarios de Go:

```bash
go install .
```

Si ese directorio está en tu `PATH`, podrás invocar `pascc` desde cualquier ubicación. Para usarlo sin compilar la herramienta primero, ejecuta los comandos con `go run .`.

## Uso

### Ejecutar un programa

`run` compila el archivo en un directorio temporal, ejecuta el resultado y elimina los archivos temporales al terminar:

```bash
./pascc run demos/hello_world.psc
```

O desde el código fuente:

```bash
go run . run demos/hello_world.psc
```

### Construir un ejecutable

`build` conserva el ejecutable y el código Go generado en `build/` por defecto. El nombre del ejecutable se toma del archivo `.psc`:

```bash
./pascc build demos/hello_world.psc
```

El resultado queda en `build/hello_world` (`build/hello_world.exe` en Windows), junto con `build/hello_world.go`. Puedes elegir otro directorio con `-o` o `--output`:

```bash
./pascc build -o dist demos/hello_world.psc
```

En los ejemplos anteriores, reemplaza `./pascc` por `pascc.exe` en Windows, o usa `go run .` para ejecutar desde el código fuente.

### Depuración

Tanto `build` como `run` aceptan `--debug` para mostrar información de las etapas del compilador, incluidos tokens, AST, árbol tipado y código Go generado:

```bash
./pascc run --debug demos/operations.psc
./pascc build --debug demos/operations.psc
```

Consulta la ayuda integrada para ver opciones y comandos:

```bash
pascc --help
pascc run --help
pascc build --help
```

### Ejemplo

Crea un archivo llamado `mayoria_edad.psc`:

```text
Algoritmo MayoriaDeEdad
    Definir edad Como Entero

    Escribir "¿Cuántos años tienes?"
    Leer edad

    Si edad >= 18 Entonces
        Escribir "Eres mayor de edad"
    SiNo
        Escribir "Eres menor de edad"
    FinSi
FinAlgoritmo
```

Ejecútalo directamente o crea un binario:

```bash
pascc run mayoria_edad.psc
pascc build mayoria_edad.psc
```

## Estado del proyecto

PSeInt tiene muchos comportamientos particulares que todavía faltan por replicar. Actualmente faltan, entre otras cosas:

- Redimensionar.
- Subprocesos, funciones y retorno de valores.
- Sintaxis flexible.
- Aliases como `Proceso` o `:=`.
- El ciclo `Para Cada` y `Repetir`.
- El operador de potencia `^`.
- Funciones como `Subcadena`.
- Modo no-scope.
- Perfiles.
- Arreglos con índice inicial 1.

## Probar los ejemplos

El directorio `demos` contiene programas para probar entrada/salida, operaciones, condicionales, ciclos y `Segun`. Por ejemplo:

```bash
pascc run demos/operations.psc
pascc run demos/ages.psc
pascc run demos/switch.psc
```

Con Just también puedes compilar la herramienta y ejecutar un ejemplo:

```bash
just build
just run demos/hello_world.psc
```

Para comprobar que los paquetes compilan:

```bash
go test ./...
```

## Contribuciones

Las contribuciones son bienvenidas, especialmente para ampliar la sintaxis compatible, mejorar los diagnósticos y añadir pruebas. Antes de hacer un cambio grande, conviene abrir un issue para acordar el alcance y evitar implementaciones duplicadas.

El uso de IA está aceptado con **moderación**.
