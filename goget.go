package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	cmd_goget = "go get -u"
)

// readFile open a fileName and return your content in a slice of lines
func readFile(fileName string) (text []string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error in reading file %s", fileName)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line_bytes, _, err := reader.ReadLine()
		if err != nil {
			return
		}
		newline := strings.TrimSpace(string(line_bytes))
		if len(newline) > 0 {
			text = append(text, newline)
		}
	}
}

// appendFile append text in the end of fileName
func appendFile(fileName string, text string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error in append in file %s", fileName)
	}
	defer file.Close()

	if _, err := file.WriteString(text + "\n"); err != nil {
		log.Fatalf("error in write in %s\n", fileName)
	}
}

// writeFile write a new file with lines slice
func writeFile(fileName string, lines []string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error in open to write in file %s", fileName)
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := file.WriteString(strings.Trim(line, " \n") + "\n"); err != nil {
			log.Fatalf("error in write in %s\n", fileName)
		}
	}
}

// isFile return true if a file exist and can be read
func isFile(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}

// updateList update a list o packages
func updateList(pkgList []string) {
	r, _ := regexp.Compile("^#")
	for _, line := range pkgList {
		if !r.MatchString(line) {
			execLine := cmd_goget + " " + line
			fmt.Printf("Runing: %s\n", execLine)
			exec.Command(execLine)
		}
	}
}

func showHelp() {
	fmt.Println("Use goget <options>")
	fmt.Println("  Where:")
	fmt.Println("   -a, -add      package to list")
	fmt.Println("   -r, -remove   remove a package from list")
	fmt.Println("   -u, -update   update packages in list")
	fmt.Println("   -c, -clean    remove duplicate and commented entry in package list")
	fmt.Println("   -h, -help     show help")
}

func cleanList(fileName string) error {
	pkgs := readFile(fileName)
	newList := []string{}
	r, _ := regexp.Compile("^#")

	fmt.Printf("Clean package list: %s\n", fileName)

	for _, pkg := range pkgs {
		if !(r.MatchString(pkg) || haveLine(newList, pkg)) {
			newList = append(newList, pkg)
		}
	}
	if err := os.Rename(fileName, fileName+".bkp"); err != nil {
		return fmt.Errorf("error in rename %s file", fileName)
	}
	writeFile(fileName, newList)
	if err := os.Remove(fileName + ".bkp"); err != nil {
		return fmt.Errorf("error in remove %s file", fileName+".bkp")
	}
	return nil
}

func haveLine(lines []string, line string) bool {
	for _, l := range lines {
		if l == line {
			return true
		}
	}
	return false
}

func removePkg(fileName string, rmpkg string) error {
	fmt.Printf("Remove package: %s\n", rmpkg)
	pkgs := readFile(fileName)
	for index, pkg := range pkgs {
		if strings.Trim(pkg, " \n") == rmpkg {
			pkgs[index] = "#" + pkg
			if err := os.Rename(fileName, fileName+".bkp"); err != nil {
				return fmt.Errorf("error in rename %s file", fileName)
			}
			writeFile(fileName, pkgs)
			if err := os.Remove(fileName + ".bkp"); err != nil {
				return fmt.Errorf("error in remove %s file", fileName+".bkp")
			}
			return nil
		}
	}
	return fmt.Errorf("package %q not found in %s", rmpkg, fileName)
}

func addNewPkg(fileName string, addPkg string) error {
	fmt.Println("Add...", addPkg)
	pkgs := readFile(fileName)
	for _, pkg := range pkgs {
		if pkg == addPkg {
			return fmt.Errorf("there is a package %q in %s", addPkg, fileName)
		}
	}
	appendFile(fileName, addPkg)
	return nil
}

// func printList(list []string) {
// 	for _, line := range list {
// 		fmt.Println(line)
// 	}
// }

func main() {
	var (
		gogetList  = os.Getenv("HOME") + "/.config/goget.list"
		addPkg     string
		remPkg     string
		updateFlag bool
		cleanFlag  bool
		helpFlag   bool
	)

	flag.StringVar(&addPkg, "add", "", "add package to list")
	flag.StringVar(&addPkg, "a", "", "add package to list (shorthand)")
	flag.StringVar(&remPkg, "remove", "", "remove a package from list")
	flag.StringVar(&remPkg, "r", "", "remove a package from list (shorthand)")
	flag.BoolVar(&updateFlag, "update", false, "update packages in list")
	flag.BoolVar(&updateFlag, "u", false, "update packages in list (shorthand)")
	flag.BoolVar(&cleanFlag, "clean", false, "remove duplicate entry in package list")
	flag.BoolVar(&cleanFlag, "c", false, "remove duplicate entry in package list (shorthand)")
	flag.BoolVar(&helpFlag, "help", false, "Show help")
	flag.BoolVar(&helpFlag, "h", false, "Show help")
	flag.Parse()

	if helpFlag {
		showHelp()
		return
	} else if updateFlag {
		if isFile(gogetList) {
			pkgs := readFile(gogetList)
			updateList(pkgs)
		} else {
			fmt.Println("List of packages not found!")
		}
		return
	} else if cleanFlag {
		cleanList(gogetList)
		return
	}

	if addPkg != "" {
		if err := addNewPkg(gogetList, addPkg); err != nil {
			fmt.Println(err.Error())
		}
	}

	if remPkg != "" {
		if err := removePkg(gogetList, remPkg); err != nil {
			fmt.Println(err.Error())
		}
	}
}
