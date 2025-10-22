package hook

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func CheckCoreDep() error {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("❌ go.mod file not found")
		return err
	}

	modData, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Printf("❌ Error reading go.mod: %v\n", err)
		return err
	}

	// Check for uncommented replace directive
	lines := bytes.Split(modData, []byte("\n"))
	for _, line := range lines {
		trimmedLine := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmedLine, []byte("replace github.com/anotik/anocore => ../anocore")) {
			fmt.Println("❌ Found uncommented local replace directive pointing to ../anocore in go.mod")
			return errors.New("found uncommented local replace directive pointing to ../anocore in go.mod")
		}
	}

	// Run go mod vendor
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "vendor")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("❌ go mod vendor failed")
		if stderr.Len() > 0 {
			fmt.Print(stderr.String())
		}
		return err
	}

	return nil
}

func RunTests(path string) error {
	cmd := exec.Command("go", "test", path)
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ go test failed")
		return err
	}
	fmt.Println("✅ go test passed successfully!")
	return nil
}

func Vet(path string) error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "vet", path)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("❌ go vet failed")
		if stderr.Len() > 0 {
			fmt.Print(stderr.String())
		}
		return err
	}

	fmt.Println("✅ go vet passed successfully!")
	return nil
}
