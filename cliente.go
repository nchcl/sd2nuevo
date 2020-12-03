 package main

 import (
         "fmt"
         "log"
         "math"
         "os"
         "strconv"
         "bufio"
         "context"
         "time"
         "math/rand"
         "strings"
         "io/ioutil"
         
         "github.com/nchcl/sd/chat"
         "google.golang.org/grpc"
 )


//Funcion para ejecutar el downloader, pide las direcciones al nameNode y luego los chunks a los dataNodes
func downloader() {
    var name string
    fmt.Println("¿Que libro desea descargar?")
    fmt.Scan(&name)
    
    conn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()
    
    c := chat.NewChatServiceClient(conn)
    response, err := c.ConsultarUbicacion(context.Background(), &chat.Signal{Nombre: name})
    if err != nil {
        log.Fatalf("Error: %s", err)
    }
    
    log.Printf("%s", response.Nombre)
    
    direcciones := strings.Split(response.Body, ",")
    partes := len(direcciones)
    
    for i:=0; i < partes; i++ {
        archivo := name+" Parte_"+strconv.Itoa(i+1)
        
        connchunk, err := grpc.Dial(direcciones[i], grpc.WithInsecure())
        if err != nil {
            log.Fatalf("did not connect: %s", err)
        }
        defer connchunk.Close()
    
        c := chat.NewChatServiceClient(connchunk)
        respuesta, err := c.PedirChunk(context.Background(), &chat.Signal{Nombre: archivo})
        if err != nil {
            log.Fatalf("Error: %s", err)
        }
        log.Printf("%s", respuesta.Name)
        ioutil.WriteFile(archivo, respuesta.Data, 0777)
    }
    
    combinar(name, uint64(partes))
}

//Funcion para subir un libro, primero checkea si los nodos estan vivos o no
//Se elige un dataNode aleatorio
func uploader(nombre_libro string) {
    var nodos_vivos [4]int = [4]int{1,1,1,1}
    var nodo_a_enviar int = azar()
    var errconn error
    
    for j := 1; j < 4; j++ {
        _, errconn = grpc.Dial(addresses[j], grpc.WithInsecure(), grpc.WithBlock(), grpc.FailOnNonTempDialError(true))
        if errconn != nil {
            fmt.Printf("Nodo %d muerto\n", j)
            nodos_vivos[j] = 0
        }
    }
    
    for nodos_vivos[nodo_a_enviar] == 0 {
        fmt.Printf("Nodo %d caido, reintentando con otro nodo...\n", nodo_a_enviar)
        if nodo_a_enviar == 3 {
            nodo_a_enviar = 1
        } else {
            nodo_a_enviar++
        }
    }
    
    var conn *grpc.ClientConn
    conn, err := grpc.Dial(addresses[nodo_a_enviar], grpc.WithInsecure())
    if err != nil {
        log.Fatalf("did not connect: %s", err)
    }
    defer conn.Close()

    fileToBeChunked := nombre_libro
    file, err := os.Open(fileToBeChunked)
    if err != nil {
            fmt.Println(err)
            os.Exit(1)
    }
    defer file.Close()

    fileInfo, _ := file.Stat()
    var fileSize int64 = fileInfo.Size()
    const fileChunk = 256000 // 1 MB, change this to your requirement
    totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
    
    for i := uint64(0); i < totalPartsNum; i++ {
        partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
        partBuffer := make([]byte, partSize)
        file.Read(partBuffer)
        
        fileName := nombre_libro+" Parte_"+strconv.FormatUint(i+1, 10)
        
        c := chat.NewChatServiceClient(conn)
        response, err := c.SendChunks(context.Background(), &chat.Chunk{Name: fileName,Parts: int32(totalPartsNum),Data: partBuffer})
        if err != nil {
            log.Fatalf("Error: %s", err)
        }
        log.Printf("%s", response.Body)
    }
    
    c := chat.NewChatServiceClient(conn)
    
    if tipo_algoritmo == 1 {
        response, err := c.TransferenciaLista(context.Background(), &chat.Signal{Id: int32(totalPartsNum)+20, Nombre: nombre_libro, Iden: int32(nodo_a_enviar)}) // aqui debo mandar el nombre del archivo...
        if err != nil {
            log.Fatalf("Error: %s", err)
        }
        log.Printf("%s", response.Body)
    } else {
        response, err := c.TransferenciaLista(context.Background(), &chat.Signal{Id: int32(totalPartsNum), Nombre: nombre_libro, Iden: int32(nodo_a_enviar)})
        if err != nil {
            log.Fatalf("Error: %s", err)
        }
        log.Printf("%s", response.Body)
    }

}


//Funcion para combinar las partes luego de descargarlas con el downloader
func combinar(nombre_libro string , cantidad_partes uint64){
        var file *os.File
        var err error
        var totalPartsNum uint64
        totalPartsNum=cantidad_partes

        newFileName := nombre_libro
        _, err = os.Create(newFileName)

        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

        file, err = os.OpenFile(newFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

        var writePosition int64 = 0

        for j := uint64(0); j < totalPartsNum; j++ {

        //read a chunk
            currentChunkFileName := nombre_libro+" Parte_"+strconv.FormatUint(j+1, 10)

            newFileChunk, err := os.Open(currentChunkFileName)

            if err != nil {
                 fmt.Println(err)
                 os.Exit(1)
            }

            defer newFileChunk.Close()

            chunkInfo, err := newFileChunk.Stat()

            if err != nil {
                 fmt.Println(err)
                 os.Exit(1)
            }

            // calculate the bytes size of each chunk
            // we are not going to rely on previous data and constant

            var chunkSize int64 = chunkInfo.Size()
            chunkBufferBytes := make([]byte, chunkSize)

            fmt.Println("Appending at position : [", writePosition, "] bytes")
            writePosition = writePosition + chunkSize

            // read into chunkBufferBytes
            reader := bufio.NewReader(newFileChunk)
            _, err = reader.Read(chunkBufferBytes)

            if err != nil {
                 fmt.Println(err)
                 os.Exit(1)
            }

            n, err := file.Write(chunkBufferBytes)

            if err != nil {
                 fmt.Println(err)
                 os.Exit(1)
            }

            file.Sync() //flush to disk

            chunkBufferBytes = nil // reset or empty our buffer

            fmt.Println("Written ", n, " bytes")

            fmt.Println("Recombining part [", j, "] into : ", newFileName)
        }

        file.Close()

}

//Funcion para elegir un numero al azar
func azar() int {
    rand.Seed(time.Now().UnixNano())
    min := 1
    max := 3
    var number int = rand.Intn(max - min + 1) + min
    return number
    
}

var addresses[4] string = [4]string{":9000",":9001",":9002",":9003"}
var tipo_algoritmo int

func main() {

        var tipo int
        var nombre_libro string
        
        fmt.Println("1. Uploader")
        fmt.Println("2. Downloader")
        fmt.Scan(&tipo)

        switch tipo {
                case 1:
                        fmt.Println("Seleccione tipo de algoritmo")
                        fmt.Println("1. Exclusion Mutua Distribuida")
                        fmt.Println("2. Exclusion Mutua Centralizada")
                        fmt.Scan(&tipo_algoritmo)
                        fmt.Println("¿Que libro desea subir?")
                        fmt.Scan(&nombre_libro)                 
                        uploader(nombre_libro)
                case 2: 
                        downloader()
        }

        
 }
