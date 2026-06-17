@ECHO OFF
for /f "delims=" %%i in ('git rev-list -1 HEAD') do set GIT_COMMIT=%%i
for /f "eol=; tokens=2 delims=^+^)" %%I in ('wmic timezone get caption /format:list') do set BUILD_TIMEZONE=%%I
for /f "delims=" %%i in ('echo %date:~6,4%-%date:~3,2%-%date:~0,2%T%time:~0,2%:%time:~3,2%:%time:~6,2%+%BUILD_TIMEZONE%') do set BUILD_TIMESTAMP=%%i
@ECHO ON
echo GIT Commit: %GIT_COMMIT%
echo Build Timestamp: %BUILD_TIMESTAMP%
go build -tags prod -ldflags "-s -w -X main.buildGitCommit=%GIT_COMMIT% -X main.buildTimestamp=%BUILD_TIMESTAMP%" -o manager.exe ../manager
go build -tags prod -ldflags "-s -w -X main.buildGitCommit=%GIT_COMMIT% -X main.buildTimestamp=%BUILD_TIMESTAMP%" -o broker.exe ../broker
go build -tags prod -ldflags "-s -w -X main.buildGitCommit=%GIT_COMMIT% -X main.buildTimestamp=%BUILD_TIMESTAMP%" -o agent.exe ../agent
go build -tags prod -ldflags "-s -w -X main.buildGitCommit=%GIT_COMMIT% -X main.buildTimestamp=%BUILD_TIMESTAMP%" -o backend.exe ../web_backend
go build -tags prod -ldflags "-s -w -X main.buildGitCommit=%GIT_COMMIT% -X main.buildTimestamp=%BUILD_TIMESTAMP%" -o importer.exe ../importer
go build -tags prod -ldflags "-s -w -X main.buildGitCommit=%GIT_COMMIT% -X main.buildTimestamp=%BUILD_TIMESTAMP%" -o pgproxy.exe ../pgproxy
set /p DONE=Hit ENTER to quit...