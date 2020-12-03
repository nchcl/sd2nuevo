# SD T2
Integrantes:
    Bernabe Garcia, rol: 201773621-6
    Ignacio Figueroa, rol: 201773526-0

Consideraciones generales:
    * Los 4 archivos principales que forman el sistema son datanode1.go (datanode 1), datanode2.go (datanode 2), datanode3.go (datanode 3), namenode.go (namenode) y  cliente.go (cliente)
    * Todas las funciones de comunicacion de grpc asi como tambien el cerebro del sistema se encuentran en el archivo chat_grpc.pb.go (dentro de la carpeta chat).
    * El archivo chat.proto posee las definiciones de los protocol buffers.
    * Para un correcto funcionamiento, el orden de ejecucion deberia ser datanode1, datanode2, datanode3, namenode y cliente.
    * El log se debe eliminar si quiero empezar un proceso nuevamente de subida y bajada, ya que asumimos que el log seria persistente y no se subiria el mismo libro 2 veces

datanode1:
    Instruccions:
        * Ingresar a la Maquina Virtual dist117
        * Ingresar a la siguiente ruta "cd datanode1/sd2nuevo/datanode1"
        * Iniciar el sistema corriendo el archivo makefile escribiendo "make" en la consola.
        * Iniciar el resto de los sistemas

    Consideraciones:

datanode2:
    Instruccions:
        * Ingresar a la Maquina Virtual dist118
        * Ingresar a la siguiente ruta "cd datanode2/sd2nuevo/datanode2"
        * Iniciar el sistema corriendo el archivo makefile escribiendo "make" en la consola.
        * Iniciar el resto de los sistemas

    Consideraciones:

datanode3:
    Instruccions:
        * Ingresar a la Maquina Virtual dist119
        * Ingresar a la siguiente ruta "cd datanode3/sd2nuevo/datanode3"
        * Iniciar el sistema corriendo el archivo makefile escribiendo "make" en la consola.
        * Iniciar el resto de los sistemas

    Consideraciones:

namenode:
    Instruccions:
        * Ingresar a la Maquina Virtual dist120
        * Ingresar a la siguiente ruta "cd namenode/sd2nuevo/"
        * Iniciar el sistema corriendo el archivo makefile escribiendo "make namenode" en la consola.

    Consideraciones:

Cliente:
    Instruccions:
        * Ingresar a la Maquina Virtual dist120
        * Ingresar a la siguiente ruta "cd namenode/sd2nuevo"
        * Iniciar el sistema corriendo el archivo makefile escribiendo "make cliente" en la consola.

    Consideraciones:
    	* Se debe ingresar el nombre del pdf con el .pdf, es decir un nombre bien apropiado seria "excelente_tarea.pdf"
