# IPC-Chat
Parcial 2 correspondiente a la materia Sistemas Operativos (SI2004) de la Universidad EAFIT.

Desarrollado por:
- Mateo Pineda Álvarez
- Esteban Álvarez Zuluaga
- Santiago Idárraga Ceballos

## Requisitos técnicos:
- Sistema Operativo Linux
- GoLang Versión 1.25

## Ejecución del programa
Para garantizar una fácil ejecución, se creó un `Makefile`, luego el proceso de ejecución va así:
- En una terminal, ejecuta make run-server para ejecutar. Allí se crean las salas de chat definidas en el código.
- Crea una terminal por cliente, y ejecuta el comando make run-client. Luego ingresa el nombre del cliente y la sala de chat a la que deseas ingresar.

¡Listo! Ya pueden empezar a enviar y recibir mensajes en las salas propuestas.

## Funcionalidades Adicionales:
- Registro de mensajes en un archivo JSON por sala. Al finalizar la ejecución se guarda un historial de los mensajes en `src/Logs/-nombre de la sala-`.
- Comando `/leave` para que un usuario cierre el programa y se eliminen correctamente las colas asociadas a él.