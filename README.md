<div align="center">

![pasc-inverted-banner](assets/pasc-inverted-banner.svg)

# `PASC`

**Compilador experimental de pseudocódigo PSeInt**

[![Go](https://img.shields.io/badge/go-1.25%2B-blue?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Pseint](https://img.shields.io/badge/compilador%20de-pseint-yellow?style=for-the-badge&logoColor=white)](https://pseint.sourceforge.net/)
[![Plataforma](https://img.shields.io/badge/multiplataforma-3c3c3c?style=for-the-badge)](#requisitos)
[![Compatibilidad](https://img.shields.io/badge/compila-100%25%20de%20programas%20v%C3%A1lidos-brightgreen?style=for-the-badge)](#compatibilidad-con-programas-reales)

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
- Arreglos con `Dimensionar` y `Redimensionar`, en base 0 o base 1 (ver [Base de los arreglos](#base-de-los-arreglos)).
- Ciclos `Para Cada elemento De arreglo`, donde la variable es una referencia al elemento.
- Expresiones en lenguaje coloquial: `x es par`, `es impar`, `es positivo`, `es cero`, `es mayor que`, `es menor o igual que`, `es distinto de`, `es divisible por`, `es multiplo de`, etc.
- Verificaciones en tiempo de compilación: conversiones de real a entero sin `Redon()`/`Trunc()` e índices fuera de rango.

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

También puedes usar `--fast` para compilar sin las dependencias de terminal. En ese modo, `Esperar Tecla` muestra un mensaje y espera a que el usuario presione Enter:

```bash
./pascc build --fast demos/hello_world.psc
```

En los ejemplos anteriores, reemplaza `./pascc` por `pascc.exe` en Windows, o usa `go run .` para ejecutar desde el código fuente.

### Base de los arreglos

PSeInt permite que los arreglos empiecen en 0 o en 1 según el perfil. Tanto `build` como `run` aceptan `--index-base`:

```bash
./pascc run --index-base 1 programa.psc   # v[1] es el primer elemento
./pascc run --index-base 0 programa.psc   # v[0] es el primer elemento
./pascc run programa.psc                  # auto (predeterminado)
```

En modo `auto`, el analizador deduce la base a partir de cómo se indexan los arreglos: un índice que puede valer `0` o un ciclo que llega hasta `n - 1` indica base 0; un índice que alcanza el tamaño del arreglo (`v[n]` con `Dimensionar v[n]`) o un ciclo `Para i <- 1 Hasta n` indica base 1. Si no hay indicios, se usa base 0, y si los indicios se contradicen se muestra una advertencia.

Con la base ya decidida, el compilador evalúa expresiones constantes y rangos de ciclos para detectar antes de ejecutar:

- índices constantes fuera de rango (error) y ciclos que se salen del arreglo (advertencia);
- dimensiones menores que 1 (error);
- conversiones de real a entero: un valor constante con decimales es un error (`k = 7 / 2`), y un valor que no se puede demostrar entero genera una advertencia que sugiere usar `Redon()` (o su alias `Redondear()`) o `Trunc()`.

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

## Compatibilidad con programas reales

Para medir la compatibilidad, `pascc` se probó contra los **18 repositorios más populares de GitHub con el tema [`pseint`](https://github.com/topics/pseint)**: **531 programas `.psc`** escritos por estudiantes y docentes, sin modificar, con codificaciones mixtas (UTF-8 y Windows-1252), estilos de sintaxis distintos y arreglos en base 0 y en base 1.

| Resultado | Programas | |
| --- | ---: | ---: |
| Compilan a un ejecutable nativo | **520 / 531** | **97.9 %** |
| Compilan, contando solo los programas válidos | **520 / 520** | **100 %** |
| Rechazados por errores reales en el código fuente | 11 | |
| Base de arreglos deducida automáticamente como 1 | 40 | |

Los 11 programas rechazados tienen errores reales que PSeInt también rechaza: archivos incompletos, plantillas sin rellenar, comillas que no cierran, `Definir a, b, Como Entero`, etc. En todos los casos `pascc` señala el archivo, la línea y la columna del problema.

**Ejecución.** Cada ejecutable se corrió respondiendo `3` a todas las lecturas. Además de los programas que terminan normalmente, se revisaron a mano todos los que no lo hicieron: **ninguno de los 520 falla por una limitación del compilador**. Todos esos casos se explican por la entrada sintética (menús que nunca reciben la opción de salida, esperas con `Esperar`) o por errores del propio programa: leer fuera de una cadena con `Subcadena`, guardar decimales en una variable `Entero` o ciclos infinitos.

**Deducción de la base de arreglos.** Con la detección automática de base 0 o base 1, **35 programas que antes fallaban en tiempo de ejecución ahora funcionan** y ninguno empeoró. Esos programas recorren arreglos con `Para i <- 1 Hasta n`. Además, las variables sin declarar que empiezan en `0` y después reciben decimales se amplían automáticamente a reales, como en PSeInt.

## Estado del proyecto

PSeInt tiene muchos comportamientos particulares y todavía quedan algunos por replicar:

- Modo no-scope.
- Opciones de perfil distintas de la base de los arreglos.

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
