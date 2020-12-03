package chat

import (
    "io/ioutil"
    "os"
    "log"
    "strings"
    "fmt"
    "strconv"
    "bufio"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Server struct {
}

//Direcciones
var addresses[4] string = [4]string{":9000",":9001",":9002",":9003"}
var nodos[] string = []string{":9001",":9002",":9003"}

func remove(s []string, i int) []string {
    s[i] = s[len(s)-1]
    return s[:len(s)-1]
}

//Funcion para enviar chunks y escribir los archivos
func (s *Server) SendChunks(ctx context.Context, in *Chunk) (*Signal, error) {
    ioutil.WriteFile(in.Name, in.Data, 0777)
    
	return &Signal{Body: "Parte lista"}, nil
}

//Funcion para enviar la propuesta del centralizado, es invocada por un dataNode y ejecutada por el nameNode
func (s *Server) EnviarPropuesta(ctx context.Context, in *Signal) (*Signal, error) {
    var nodos_muertos [4]int = [4]int{1,1,1,1}
    var errconn error
    var reformular_propuesta bool = false
    var nodo_distribuidor int = int(in.Id)
    
    for i := 1; i < 4; i++ {
        if i == nodo_distribuidor {
            continue
        }
        
        _, errconn = grpc.Dial(addresses[i], grpc.WithInsecure(), grpc.WithBlock(), grpc.FailOnNonTempDialError(true))
        if errconn != nil {
            fmt.Printf("Nodo %d muerto\n", i)
            nodos_muertos[i] = 0
            reformular_propuesta = true
        }
    }
    
    //LOG HERE
    var partes int = len(strings.Split(in.Body, ","))
    var prop string
    var prop_split []string
    if reformular_propuesta {
        prop = nueva_propuesta(partes, nodos_muertos)
        prop_split  = strings.Split(prop, ",")
    } else {
        prop_split = strings.Split(in.Body, ",")
        prop = in.Body
    }
    write_log(in.Nombre, partes, prop_split)
    
    return &Signal{Body: prop, Id: int32(nodo_distribuidor)}, nil
}

//Funcion para enviar la propuesta del descentralizado
func (s *Server) EnviarPropuestaDist(ctx context.Context, in *Signal) (*Signal, error) {
    
    var propuesta = in.Body
    log.Printf("Es la Propuesta en EnviarPropuesta: %s", propuesta) //propuesta
    var conn *grpc.ClientConn
    conn, err := grpc.Dial(addresses[int(in.Id)], grpc.WithInsecure())
    if err != nil {
        log.Fatalf("did not connect: %s", err)
    }
    defer conn.Close()
    
    c := NewChatServiceClient(conn)
    response, err := c.RecibirPropuestaDist(context.Background(), &Signal{Response: true, Nombre:in.Nombre, Body: in.Body, Id: in.Id})
    if err != nil {
    log.Fatalf("Error: %s", err)
    }
    log.Printf("%s", response.Body)
    
    return &Signal{Body: "Chunks distribuidos"}, nil
}

//Funcion para escribir el log, la invoca un dataNode y la ejecuta el nameNode
func (s *Server) EscribirLog(ctx context.Context, in *Signal) (*Signal, error) {
    var nombre string = in.Nombre
    var partes int = int(in.Id)
    var direcciones = strings.Split(in.Body, ",")
    
    f, err := os.OpenFile("log.txt",os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Println(err)
    }
    defer f.Close()
    
    var buffer string = nombre+" "+strconv.Itoa(partes)+"\n"
    
    if _, err := f.WriteString(buffer); err != nil {
        log.Println(err)
    }
    
    for i := 0; i < partes; i++ {
        if err != nil {
            log.Fatalf("fail: %s", err)
        }
        
        buffer = nombre+" Parte_"+strconv.Itoa(i+1)+" "+direcciones[i]+"\n"
        
        if _, err := f.WriteString(buffer); err != nil {
            log.Println(err)
        }
    }   


    return &Signal{Body: "Log actualizado"}, nil
}

//Funcion para recibir una propuesta del algoritmo descentralizado, la ejecuta un dataNode
func (s *Server) RecibirPropuestaDist(ctx context.Context, in *Signal) (*Signal, error) {
    log.Printf("Propuesta aprobada")
    var propuesta = strings.Split(in.Body, ",")
    var current_node = int(in.Id)
    
    var conn1 *grpc.ClientConn
    conn1, err := grpc.Dial(addresses[0], grpc.WithInsecure())
    if err != nil {
        log.Fatalf("did not connect: %s", err)
    }
    defer conn1.Close()
        
    cs := NewChatServiceClient(conn1)
    response, err := cs.EscribirLog(context.Background(), &Signal{Nombre:in.Nombre, Body: in.Body, Id: int32(len(propuesta))})
    if err != nil {
    log.Fatalf("Error: %s", err)
    }
    log.Printf("%s", response.Body)
        
    
    for i := 0; i < len(propuesta); i++ {
        if propuesta[i] == addresses[current_node] {
            continue
        }
        
        chunkName := in.Nombre +" Parte_"+strconv.Itoa(i+1)
        file, err := os.Open(chunkName)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        defer file.Close()
        
        fileInfo, _ := file.Stat()
        var fileSize int64 = fileInfo.Size()
        buffer := make([]byte, fileSize)
        file.Read(buffer)
        var conn *grpc.ClientConn
        conn, errconn := grpc.Dial(propuesta[i], grpc.WithInsecure())
        if errconn != nil {
            log.Fatalf("did not connect: %s", err)
        }
        defer conn.Close()
        
        c := NewChatServiceClient(conn)
        response, err := c.SendChunks(context.Background(), &Chunk{Name: chunkName,Data: buffer})
        if err != nil {
	       log.Fatalf("Error: %s", err)
        }
        log.Printf("%s", response.Body)
        
        err = os.Remove(chunkName) 
        if err != nil { 
            log.Fatal(err) 
        }
    }
    
    
	return &Signal{Body: "Chunks distribuidos"}, nil
}

//Funcion para generar una primera propuesta inicial para el algoritmo centralizado
func generar_propuesta(partes int) string {
    var propuesta string
    var nodo int = 1
    for i := 0; i < partes; i++ {
        if i == partes-1 {
            propuesta += strconv.Itoa(nodo)
            break
        }
        
        propuesta += strconv.Itoa(nodo)+","
        if nodo == 3 {
            nodo = 1
        } else {
            nodo++
        }
    }
    
    return propuesta
}

//Funcion para consultar las ubicaciones de las partes de un libro, la invoca el cliente y la ejecuta el nameNode
func (s *Server) ConsultarUbicacion(ctx context.Context, in *Signal) (*Signal, error) {
    
    var name string = in.Nombre
    var direcciones []string
    
    file, err := os.Open("log.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        romper_linea := strings.Fields(scanner.Text())
        if romper_linea[0] == name {
            partes,_ := strconv.Atoi(romper_linea[1])
            
            for i := 0; i < partes; i++ {
                scanner.Scan()
                linea := strings.Fields(scanner.Text())
                direcciones = append(direcciones, linea[2])
            }
            break
        }
    }
    
    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

    return &Signal{Nombre: "Direcciones encontradas", Body: strings.Join(direcciones[:],",")}, nil
}

//Funcion para pedir un chunk a un dataNode
func (s *Server) PedirChunk(ctx context.Context, in *Signal) (*Chunk, error) {
    
    name := in.Nombre
    file, err := os.Open(name)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer file.Close()
    
    fileInfo, _ := file.Stat()
    var fileSize int64 = fileInfo.Size()
    buffer := make([]byte, fileSize)
    file.Read(buffer)
        
	return &Chunk{Name: "Chunk recibido", Data: buffer}, nil
}

//Funcion que avisa que se finalizo el traspaso de los chunks
func (s *Server) TransferenciaLista(ctx context.Context, in *Signal) (*Signal, error) {
    log.Printf("%s", in.Nombre)
    //Hardcodeare con la suma de 20 cuando sea Distribuido.
    if in.Id>20 {
        var nodos_disponibles []string
        var nodos_disponibles_aux []string
        
        var propuesta_generada string
        
        nodos_disponibles=nodos
        log.Printf("nodos_disponibles: %s",nodos_disponibles)
        propuesta_generada = generar_propuesta_dist(int(in.Id), nodos_disponibles)
        log.Printf("nodos_disponibles: %s",nodos_disponibles)
        nodos_disponibles_aux = revision_propuesta_dist(propuesta_generada, int(in.Id), nodos_disponibles)
        log.Printf("nodos_disponibles: %s",nodos_disponibles)
        log.Printf("nodos_disponibles_aux: %s",nodos_disponibles_aux)
        for len(nodos_disponibles_aux)!=len(nodos_disponibles) {
            nodos_disponibles = nodos_disponibles_aux
            propuesta_generada= generar_propuesta_dist(int(in.Id),nodos_disponibles)
            nodos_disponibles_aux= revision_propuesta_dist(propuesta_generada, int(in.Id), nodos_disponibles)
        }
        log.Printf("Envio la Propuesta")
        enviar_propuesta_dist(propuesta_generada, int(in.Id), nodos_disponibles, in.Nombre, int(in.Iden))

        return &Signal{Body: "Transferencia terminada"}, nil
    }
    
    var name string = in.Nombre
    var propuesta string = generar_propuesta(int(in.Id))
    var conn *grpc.ClientConn
	conn, err := grpc.Dial(addresses[0], grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()
    
    c := NewChatServiceClient(conn)
    response, err := c.EnviarPropuesta(context.Background(), &Signal{Id: int32(in.Iden), Body: propuesta, Nombre: name})
    if err != nil {
        log.Fatalf("Error: %s", err)
    }
    
    go enviar_propuesta_nodos(response.Body, name, int(in.Iden))
    
    return &Signal{Body: "Transferencia terminada"}, nil
}

//Funcion de prueba para enviar un ping
func (s *Server) SendMensaje(ctx context.Context, in *Alerta) (*Alerta, error) {
    log.Printf("Mensaje recibido del cliente: %s", in.Mensaje)
    return &Alerta{Mensaje: "Bien"}, nil
}

//Funcion del algoritmo centralizado para mandar la propuesta aprobada a los otros dataNodes
func enviar_propuesta_nodos(prop string, name string, id_nodo int) {
    
    var current_node = strconv.Itoa(id_nodo)
    var propuesta = strings.Split(prop, ",")
    
    for i := 0; i < len(propuesta); i++ {
        if propuesta[i] == current_node {
            continue
        }
        
        chunkName := name+" Parte_"+strconv.Itoa(i+1)
        fmt.Println(chunkName)
        file, err := os.Open(chunkName)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        defer file.Close()
        
        fileInfo, _ := file.Stat()
        var fileSize int64 = fileInfo.Size()
        buffer := make([]byte, fileSize)
        file.Read(buffer)
        
        index, err := strconv.Atoi(propuesta[i])
        if err != nil {
            log.Fatalf("fail: %s", err)
        }
        
        var conn *grpc.ClientConn
        conn, errconn := grpc.Dial(addresses[index], grpc.WithInsecure())
        if errconn != nil {
            log.Fatalf("did not connect: %s", err)
        }
        defer conn.Close()
        
        c := NewChatServiceClient(conn)
        response, err := c.SendChunks(context.Background(), &Chunk{Name: chunkName,Data: buffer})
        if err != nil {
            log.Fatalf("Error: %s", err)
        }
        log.Printf("%s", response.Body)
        
        err = os.Remove(chunkName) 
        if err != nil { 
            log.Fatal(err) 
        } 
    }
}

//Funcion para generar una propuesta del algoritmo descentralizado
func generar_propuesta_dist(partes int, nodos_disponibles []string) string {
    partes=partes - 20 // para conseguir las partes reales.
    var propuesta string
    var nodo int = 0
    for i := 0; i < partes; i++ {
        if i == partes-1 {
            propuesta += nodos_disponibles[nodo]
            break
        }
    
        propuesta += nodos_disponibles[nodo]+","
        if nodo == len(nodos_disponibles)-1 {
            nodo = 0
        } else {
            nodo++
        }
    }   
    return propuesta
}


//Funcion para revisar la propuesta del algoritmo descentralizado
func revision_propuesta_dist(propuesta string, partes int, nodos_disponibles []string) []string {

    for i:=0; i<len(nodos_disponibles); i++{

        conn_revision, err_con := grpc.Dial(nodos_disponibles[i], grpc.WithInsecure())
        //Mensaje a cada 1 de las maquinas y recibir algo de vuelta para saber si esta funcionando.
        if err_con != nil {
            log.Fatalf("did not connect: %s", err_con)
        }

        defer conn_revision.Close()
        c := NewChatServiceClient(conn_revision)
        response, err_con := c.SendMensaje(context.Background(), &Alerta{Mensaje: "Mensaje de Prueba"})
        if err_con != nil {
            nodos_disponibles=remove(nodos_disponibles, i)
            return nodos_disponibles
        }

        log.Printf("%s", response.Mensaje)
    } //Tengo que borrar el nodo inmediatamente? o reviso todos y dsps elimino? 
    return nodos_disponibles
}

//Funcion para enviar la propuesta del algoritmo descentralizado
func enviar_propuesta_dist(propuesta string, partes int, nodos_disponibles []string, nombre string, id_nodo int) {
    var conn *grpc.ClientConn
    conn, err := grpc.Dial(nodos_disponibles[0], grpc.WithInsecure())
    if err != nil {
        log.Printf("did not connect: %s", err)
    }
    defer conn.Close()
    c := NewChatServiceClient(conn)
    response, err := c.EnviarPropuestaDist(context.Background(), &Signal{Id: int32(id_nodo), Body: propuesta, Nombre: nombre})
    if err != nil {
        log.Printf("Error: %s", err)
    }
    log.Printf("%s", response.Body)
}

//Funcion local ejecutada por el nameNode para escribir en el log, parte del algoritmo centralizado
func write_log(nombre string, partes int, nodos []string) {
    f, err := os.OpenFile("log.txt",os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Println(err)
    }
    defer f.Close()
    
    var buffer string = nombre+" "+strconv.Itoa(partes)+"\n"
    
    if _, err := f.WriteString(buffer); err != nil {
        log.Println(err)
    }
    
    for i := 0; i < partes; i++ {
        index, err := strconv.Atoi(nodos[i])
        if err != nil {
            log.Fatalf("fail: %s", err)
        }
        
        buffer = nombre+" Parte_"+strconv.Itoa(i+1)+" "+addresses[index]+"\n"
        
        if _, err := f.WriteString(buffer); err != nil {
            log.Println(err)
        }
    }   

}

//Generacion de una nueva propuesta considerando los nodos muertos
func nueva_propuesta(partes int, nodos_muertos [4]int) string {
    var propuesta string
    var nodo int = 1
    for i:= 0; i < partes; i++ {
        for nodos_muertos[nodo] == 0 {
            nodo++
            if nodo == 4 {
                nodo = 1
            }
        }
        
        if i == partes-1 {
            propuesta += strconv.Itoa(nodo)
            break
        }
        
        propuesta += strconv.Itoa(nodo)+","
        if nodo == 3 {
            nodo = 1
        } else {
            nodo++
        }
    }
    
    return propuesta
}

