// this package creates static folder, creates files: index.html, virtual.conf
package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/daominah/livestream/zconfig"
)

func main() {
	err := os.Mkdir(zconfig.StaticFolder, 0777)
	if err != nil {
		fmt.Println("Mkdir", err)
	}
	fmt.Println("Created folder: ", zconfig.StaticFolder)
	//
	indexHtml := `
	<html>
      <head>
        <title>DaoMinAh</title>
      </head>
      <body>
        Hohohaha
      </body>
    </html>`
	err = ioutil.WriteFile(
		zconfig.StaticFolder+"/index.html", []byte(indexHtml), 0777)
	fmt.Println("Created index.html")
	//
	if err != nil {
		fmt.Println("WriteFile index.html ", err)
	}
	virtualConf := fmt.Sprintf(
		`#
# A virtual host using mix of IP-, name-, and port-based configuration
#

server {
    listen       %v;
    server_name  example.com;

    location / {
        root   %v;
        index  index.html;
    }
}`, zconfig.StaticDownloadPort, zconfig.StaticFolder)
	err = ioutil.WriteFile(
		zconfig.StaticFolder+"/virtual.conf", []byte(virtualConf), 0777)
	if err != nil {
		fmt.Println("WriteFile virtual.conf ", err)
	}
	fmt.Println(`Created virtual.conf.
You need to copy it to /etc/nginx/conf.d/virtual.conf and /etc/init.d/nginx restart`)
}
