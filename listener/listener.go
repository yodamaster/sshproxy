// ./listener -listen 0.0.0.0:2022 ./next-program -conn_fd '{}'
//
// The point of listener is so that `next-program` doesn't have to
// handle multiple connections, and can exit on errors.
// Go doesn't "do" fork within a binary for these purposes.
package main

/*
 *  Copyright (C) 2014 Thomas Habets <thomas@habets.se>
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
 *
 */
import (
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
)

var (
	listen = flag.String("listen", "", "Listen address.")
)

func handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	args := flag.Args()[1:]
	for n := range args {
		if args[n] == "{}" {
			args[n] = "3"
		}
	}
	cmd := exec.Command(flag.Args()[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	func() {
		if f, err := conn.File(); err != nil {
			log.Printf("Command start failed: %v", err)
		} else {
			defer conn.Close()
			defer f.Close()
			cmd.ExtraFiles = []*os.File{f}
		}
		if err := cmd.Start(); err != nil {
			log.Printf("Command start failed: %v", err)
		}
	}()
	if err := cmd.Wait(); err != nil {
		log.Printf("Command wait failed: %v", err)
	}
}

func main() {
	flag.Parse()
	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("Failed to listen to %q: %v", *listen, err)
	}
	log.Printf("Ready")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept(): %v", err)
			continue
		}
		go handleConnection(conn.(*net.TCPConn))
	}
}
