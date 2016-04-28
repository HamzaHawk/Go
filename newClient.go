package main

import "net"
import "fmt"
import "bufio"
import "os"
import "strings"
import "io"
import "github.com/nyarlabo/go-crypt"
//import "reflect"

func main() {

	conn, _ := net.Dial("tcp", "127.0.0.1:3000")
	bufc := bufio.NewReader(conn)
	fmt.Fprintf(conn,"client" + "\n")
	line, _ := bufc.ReadString('\n')	
	fmt.Print("Message from server: "+line)
	fmt.Print("Enter Password: \n")
	go Read(conn)
	
	Write(conn)
	
}
func Read(conn net.Conn){
		
	for{
		bufc := bufio.NewReader(conn)
		line, err := bufc.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Print("Message from server: "+line)
		fmt.Print("Enter Password: \n")
	}

}

func Write(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		//fmt.Print("Enter Password: \n")
		Password, _ := reader.ReadString('\n')
		Password=strings.TrimSpace(Password)
		fmt.Println(len(Password))
		cipher:=crypt.Crypt(Password, "")
		fmt.Println(cipher)
		// send to socket
		var newString string=cipher
		io.WriteString(conn, newString+ "\n")

		//fmt.Fprintf(conn,newString+ "\n")
	}
}