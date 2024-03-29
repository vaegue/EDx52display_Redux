go clean

Remove-Item  .\EDx52Display -Force -Recurse -ErrorAction SilentlyContinue
Remove-Item .\Release.zip -ErrorAction SilentlyContinue

New-Item -ItemType "directory" -Path ".\EDx52Display"

go build

Copy-Item -Path EDx52display_Redux.exe,conf.yaml,LICENSE,README.md,names,DepInclude -Destination .\EDx52Display -Recurse
Rename-Item -Path .\EDx52Display\EDx52display_Redux.exe -NewName EDx52display.exe

Compress-Archive .\EDx52Display\* Release.zip