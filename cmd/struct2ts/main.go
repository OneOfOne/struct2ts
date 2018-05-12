package main

import (
	"os"

	"github.com/OneOfOne/struct2ts"
	"github.com/PathDNA/missionControl/users"
)

func main() {
	s := struct2ts.New(nil)
	s.Add(users.User{})
	s.Add(users.Login{})
	s.RenderTo(os.Stdout)
	// j, _ := json.MarshalIndent(s.structs, "", "\t")
	// fmt.Println(string(j))
}
