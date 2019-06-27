# # Uncomment if you want to debug this script but it makes the compression step really slow
# Set-PSDebug -Trace 1
echo "Installing Chocolatey Package Manager..."
Set-ExecutionPolicy Bypass -Scope Process -Force; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
echo "Installing make..."
choco install make -y

# create new directories inside of the container
New-Item -ItemType Directory -Force -Path "C:\gopath\src\github.com\influxdata"
New-Item -ItemType Directory -Force -Path "C:\output"

echo "COPYING C:\src to C:\gopath\src\github.com\influxdata\telegraf"
cp -r 'C:\src' 'C:\gopath\src\github.com\influxdata\telegraf'

# change to build directory
cd 'C:\gopath\src\github.com\influxdata\telegraf'

echo "Cleaning Build Directory..."
make clean

echo "Restoring Dependencies..."
make deps

echo "Linting Telegraf..."
make lint

echo "Testing Telegraf..."
make test-windows

echo "Making Telegraf..."
make
ls

echo "Archiving Telegraf..."
Compress-Archive -Path C:\gopath\src\github.com\influxdata\telegraf\telegraf.exe -CompressionLevel Fastest -DestinationPath C:\output\Windows-x86_64.zip -Force
ls C:\output\
echo "Done!"
