go build -tags prod -ldflags "-s -w" -o manager.exe ../manager
go build -tags prod -ldflags "-s -w" -o broker.exe ../broker
go build -tags prod -ldflags "-s -w" -o agent.exe ../agent
go build -tags prod -ldflags "-s -w" -o backend.exe ../web_backend
go build -tags prod -ldflags "-s -w" -o importer.exe ../importer
set /p DONE=Hit ENTER to quit...