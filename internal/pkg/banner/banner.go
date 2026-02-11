package banner

import (
	"fmt"
	"runtime"
)

const banner = `
   _____ _                 _        ______                     
  / ____(_)               | |      |  ____|                    
 | (___  _ _ __ ___  _ __ | | ___  | |__  __  ____ _ _ __ ___  
  \___ \| | '_ ' _ \| '_ \| |/ _ \ |  __| \ \/ / _' | '_ ' _ \ 
  ____) | | | | | | | |_) | |  __/ | |____ >  < (_| | | | | | |
 |_____/|_|_| |_| |_| .__/|_|\___| |______/_/\_\__,_|_| |_| |_|
                    | |                                         
                    |_|                                         
`

// Print 打印启动横幅，包含版本信息和构建信息
func Print(version, commitHash, buildTime string) {
	fmt.Print(banner)
	fmt.Printf("  Version:     %s\n", version)
	
	if commitHash != "" && commitHash != "unknown" {
		// 如果 commit hash 太长，只显示前 7 位
		if len(commitHash) > 7 {
			commitHash = commitHash[:7]
		}
		fmt.Printf("  Commit:      %s\n", commitHash)
	}
	
	if buildTime != "" && buildTime != "unknown" {
		fmt.Printf("  Build Time:  %s\n", buildTime)
	}
	
	fmt.Printf("  Go Version:  %s\n", runtime.Version())
	fmt.Printf("  OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
}
