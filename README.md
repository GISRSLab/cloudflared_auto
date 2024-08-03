# Cloudflare tunnel rdp management tool

This tool is made for a specific need of automatically registering service on windows.  
According to official document, once you configure a rdp tunnel in public hostname, you need to connect it as user with `cloudflared` client software, this tool helps those who are not familiar with command line and who want to run it backend.  

Have a good trip!

## Usage

You can use it from source:
```sh
git clone https://github.com/DSYZayn/cloudflared_auto
cd ./cloudflared_auto/cmd
go build -o 'auto register rdp.exe'
```
Then you get a excutable file `auto register rdp.exe`, run it as an administrator.

Or you can download from release.