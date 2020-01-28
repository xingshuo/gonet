rm client/client.exe server/server.exe

cd client && go build -o client.exe main.go
cd ..
cd server && go build -o server.exe main.go
cd ..
echo build done.