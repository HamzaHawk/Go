package main

import "net"
import "fmt"
import "bufio"
import "io"
import "myproject/mylibs"
import "github.com/nyarlabo/go-crypt"
import "strings"
import "os"

func main() {

	conn, _ := net.Dial("tcp", "127.0.0.1:3000")
	bufc := bufio.NewReader(conn)
	//line, _ := bufc.ReadString('\n')

	//fmt.Print("Message from server: "+line)
	fmt.Fprintf(conn,"slave" + "\n")
	for i := 0; i < 2; i++ {
	line, _ := bufc.ReadString('\n')	
	fmt.Print("Message from server: "+line)
}
	for{
	temp:=Read(conn)
	Write(temp,conn)
}
}
func Read(conn net.Conn) string{
		
		bufc := bufio.NewReader(conn)
		cipher, err := bufc.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cipher=strings.TrimSpace(cipher)
		fmt.Print("Message from client: "+cipher)
		bd := BruteDict.New(true, false, 1,5)
		defer bd.Close()
		for {
			id := bd.Id()
			if (decrypt(id,cipher)){
				fmt.Println("Password is :"+id)
				return id;
			}
			fmt.Println(id)
			if id == "" {
				return ""
			}
		}
		
return ""
}

func Write(temp string,conn net.Conn) {
	if temp==""{
		fmt.Print("No Password found: \n")
		io.WriteString(conn,"No Password found"+ "\n")
	}else{
		fmt.Print("Text to send: \n"+temp)
		// send to socket
		io.WriteString(conn, "Password found: "+temp+ "\n")
		//fmt.Fprintf(conn,text + "\n")
	}
		
}

func decrypt(alphabets string,cipher string) bool{
	listAlpha:=alphabets
	if ret := crypt.Crypt(listAlpha, ""); ret != cipher {
		//log.Printf(`result of Crypt is mismatch: %+v`, []byte(ret))
		return false
	} else if ret := crypt.Crypt(listAlpha, ""); ret == cipher{
		fmt.Println("Password cracked")
		return true
	}
	return false
}