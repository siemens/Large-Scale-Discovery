go build -tags prod -ldflags "-s -w" -o manager.bin ../manager
go build -tags prod -ldflags "-s -w" -o broker.bin ../broker
go build -tags prod -ldflags "-s -w" -o agent.bin ../agent
go build -tags prod -ldflags "-s -w" -o backend.bin ../web_backend
go build -tags prod -ldflags "-s -w" -o importer.bin ../importer
set /p DONE=Hit ENTER to quit...