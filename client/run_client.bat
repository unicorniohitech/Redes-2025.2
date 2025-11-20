@echo off
REM Usage: check-and-run-client.bat [HOST] [PORT]
REM Defaults: HOST=localhost PORT=8000
SETLOCAL
SET "HOST=%~1"
IF "%HOST%"=="" SET "HOST=localhost"
SET "PORT=%~2"
IF "%PORT%"=="" SET "PORT=8000"

echo Using HOST=%HOST% PORT=%PORT%

REM Minimal Go check: run 'go version' and verify it returns output
SET "GOVER="
FOR /F "delims=" %%V IN ('go version 2^>nul') DO SET "GOVER=%%V"
IF NOT DEFINED GOVER (
  echo Error: 'go version' returned no output. Install Go and ensure 'go' is in PATH.
  ENDLOCAL
  EXIT /b 1
)

echo Go found: %GOVER%
echo Running client locally with HOST=%HOST% PORT=%PORT%

REM Move to repository root (assumes script is in client\)
PUSHD "%~dp0.." >nul || (
  echo Failed to change directory to repository root.
  ENDLOCAL
  exit /b 1
)

REM Prefer a built binary if present
IF EXIST "tcp\bin\tcp.exe" (
  echo Found tcp\bin\tcp.exe - running it
  tcp\bin\tcp.exe -mode=client -address=%HOST% -port=%PORT%
  POPD >nul
  ENDLOCAL
  EXIT /b 0
)

go build -o tcp\bin\tcp.exe ..\tcp
IF ERRORLEVEL 1 (
  echo Error: Failed to build tcp client binary.
  POPD >nul
  ENDLOCAL
  EXIT /b 1
)

REM Use 'go run' to execute the client
go run ..\tcp\bin\tcp.exe -mode=client -address=%HOST% -port=%PORT%
SET "RC=%ERRORLEVEL%"
POPD >nul
ENDLOCAL
EXIT /b %RC%
