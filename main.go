package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	requests "github.com/jochasinga/requests"
)

// this function verify if a file is empty
func isEmpty(path string) (bool, error) {
    info, err := os.Stat(path)
    if err != nil {
        return false, err
    }
    return info.Size() == 0, nil
}

func main() {
    var worldlistPath string
    var hostName string
    var openDirectories []string

    // defining flags
    flag.StringVar(&worldlistPath, "w", "/home/santtos/Desktop/pentesting/programas/golang/subdomain_enumerator/wordlist/default_wordlist.txt", "Determine the wordlist, if not passed, uses a default wordlist")
    flag.StringVar(&hostName, "h", "", "Determine a hostname")
    flag.Parse()

    // if hostname not passed
    if hostName == "" {
        fmt.Println("Error: hostname is required. Use the -h flag to specify it.")
        os.Exit(1)
    }

    // chacking if file is empty
    empty, err := isEmpty(worldlistPath)
    if err != nil {
        fmt.Println("Error file validation")
    }
    if empty {
        fmt.Println("File is empty!")
        os.Exit(1)
    }

    if worldlistPath != "" {
        // open file
        file, err := os.Open(worldlistPath)
        if err != nil {
            fmt.Println("Wordlist does not exist or invalid path!")
            os.Exit(1)
        }
        defer file.Close()

        // read file
        var wordlist []string
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            wordlist = append(wordlist, scanner.Text())
        }
        if err != nil {
            fmt.Println("Scanning error!")
        }

        var wg sync.WaitGroup
        mutex := &sync.Mutex{}

        now := time.Now()
        color.New(color.FgHiRed, color.Bold).Printf("START TIME: %v\n", now.Format("15:04:05"))
        color.New(color.FgHiRed, color.Bold).Printf("TARGET: %v\n", hostName)
        color.New(color.FgHiRed, color.Bold).Printf("WORDLIST PATH: %v\n", worldlistPath)
        time.Sleep(2*time.Second)
        fmt.Printf("%-40s | %-6s\n", "URL", "Status")
	    fmt.Println(strings.Repeat("-", 50))

        // brute force
        for _, dirs := range wordlist {
            wg.Add(1)

            go func (dirs string) {
                defer wg.Done()

                req := hostName + "/" + dirs
                resp, err := requests.Get(req)
                time.Sleep(10 * time.Second)
                if err != nil {
                    fmt.Println("Error on Get ", err)
                    return
                }

                if resp.StatusCode == 200 {
                    mutex.Lock()
                    openDirectories = append(openDirectories, dirs)
                    mutex.Unlock()
                }
                if resp.StatusCode != 404 {
                    fmt.Printf("%-40s | %-6d\n", req, resp.StatusCode)
                }
            }(dirs)
        }

        wg.Wait()

    } else {
        // open file
        file, err := os.Open(worldlistPath)
        if err != nil {
            fmt.Println("Wordlist does not exist")
            return
        }
        defer file.Close()

        // read file
        var wordlist []string
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            wordlist = append(wordlist, scanner.Text())
        }

        var wg sync.WaitGroup
        mutex := &sync.Mutex{}

        // brute force
        for _, dirs := range wordlist {
            wg.Add(1)

            go func (dirs string) {
                defer wg.Done()

                req := hostName + "/" + dirs
                resp, err := requests.Get(req)
                if err != nil {
                    fmt.Println("Error on Get ", err)
                    return
                }

                if resp.StatusCode == 200 {
                    mutex.Lock()
                    openDirectories = append(openDirectories, dirs)
                    mutex.Unlock()
                }
                if resp.StatusCode != 404 {
                    fmt.Printf("%-40s | %-6d\n", req, resp.StatusCode)
                }
            }(dirs)
        }

        wg.Wait()

    }
    // display all status 200 directories
    if len(openDirectories) > 0 {
        color.New(color.FgHiGreen, color.Bold).Println("\nOpen Directories:")
        for _, dir := range openDirectories {
            fmt.Printf(" - %v/%s\n", hostName, dir)
        }
    } else {
        color.New(color.FgHiRed, color.Bold).Println("\nNo open directories found.")
    }
}