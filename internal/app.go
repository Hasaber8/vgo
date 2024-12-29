package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"text/template"

	tmpl "github.com/vg006/vgo/internal/templates"
)

var (
	errChan = make(chan error, 2)
)

func (p *Project) ScaffoldProject() error {
	// Creates the project root directory
	err := os.Mkdir(p.Name, 0754)
	if err != nil {
		return err
	}
	// -----------------------------------------------------------------
	// Changes the directory to the project root
	err = os.Chdir(p.Name)
	if err != nil {
		return err
	}
	// Initiates the project module
	_, err = exec.Command("go", "mod", "init", p.ModName).Output()
	if err != nil {
		return err
	}
	// Creates a Readme.md File
	f, err := os.Create("Readme.md")
	if err != nil {
		return err
	}
	defer f.Close()

	// Creates other project directories
	// -----------------------------------------------------------------
	// Creates the cmd directory
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		p.CreateCmdDir()
	}()
	// Creates the internal directory
	go func() {
		defer wg.Done()
		p.CreateInternalDir()
	}()

	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Importing the packages
	// -----------------------------------------------------------------
	_, err = exec.Command("go", "mod", "tidy").Output()
	if err != nil {
		return err
	}

	return nil
}

func (p *Project) CreateCmdDir() {
	// Creates the cmd directory
	err := os.Mkdir("cmd", 0754)
	if err != nil {
		errChan <- err
	}

	// Creates the server directory
	serverPath := filepath.Join("cmd", "server")
	err = os.MkdirAll(serverPath, 0754)
	if err != nil {
		errChan <- err
	}

	// Creates the server.go
	serverFile, err := os.Create(filepath.Join(serverPath, "server.go"))
	if err != nil {
		errChan <- err
	}
	defer serverFile.Close()

	// Writes into serverFile
	err = template.
		Must(
			template.
				New("server.go").
				Funcs(Functions).
				Parse(tmpl.ServerTmpl)).
		Execute(serverFile, p)
	if err != nil {
		errChan <- err
	}
}

func (p *Project) CreateInternalDir() {
	// internal
	// -----------------------------------------------------------------
	// Creates the internal directory
	err := os.Mkdir("internal", 0754)
	if err != nil {
		errChan <- err
	}

	// internal/app
	// -----------------------------------------------------------------
	// Creates the app directory
	appPath := filepath.Join("internal", "app")
	err = os.MkdirAll(appPath, 0754)
	if err != nil {
		errChan <- err
	}
	// Creates the app.go
	appFile, err := os.Create(filepath.Join(appPath, "app.go"))
	if err != nil {
		errChan <- err
	}
	defer appFile.Close()
	// Writes into appFile
	err = template.
		Must(
			template.
				New("app.go").
				Funcs(Functions).
				Parse(tmpl.AppTmpl)).
		Execute(appFile, p)
	if err != nil {
		errChan <- err
	}

	// internal/database
	// -----------------------------------------------------------------
	// Creates the database directory
	if p.Database != "none" {
		dbPath := filepath.Join("internal", "database")
		err = os.MkdirAll(dbPath, 0754)
		if err != nil {
			errChan <- err
		}
		// Creates the database.go
		dbFile, err := os.Create(filepath.Join(dbPath, "database.go"))
		if err != nil {
			errChan <- err
		}
		defer dbFile.Close()
		// Writes into dbFile
		err = template.
			Must(
				template.
					New("database.go").
					Funcs(Functions).
					Parse(tmpl.DatabaseTmpl(p.Database))).
			Execute(dbFile, p)
		if err != nil {
			errChan <- err
		}
	}

	// internal/handlers
	// -----------------------------------------------------------------
	// Creates the handlers directory
	handlersPath := filepath.Join("internal", "handlers")
	err = os.MkdirAll(handlersPath, 0754)
	if err != nil {
		errChan <- err
	}
	// Creates the handlers.go
	handlersFile, err := os.Create(filepath.Join(handlersPath, "handlers.go"))
	if err != nil {
		errChan <- err
	}
	defer handlersFile.Close()
	// Writes into handlersFile
	err = template.
		Must(
			template.
				New("handlers.go").
				Funcs(Functions).
				Parse(tmpl.HandlerTmpl(p.FrameWork))).
		Execute(handlersFile, p)
	if err != nil {
		errChan <- err
	}
	// -----------------------------------------------------------------
}

var Functions = template.FuncMap{
	"returnModName": func(p *Project) string {
		return p.ModName
	},
	"returnDbInstance": func(db string) string {
		switch db {
		case "mongo":
			return "db := db.New(a.DB_MONGODB_URI, a.DB_NAME, a.DB_COLLECTION)"
		case "none":
			return ""
		default:
			return "db := db.New(a.DB_CONNECTION_STRING)"
		}
	},
}
