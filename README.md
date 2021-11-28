# libgen-tui
A terminal based application for LibGen

[![Go Report Card](https://goreportcard.com/badge/github.com/audstanley/libgen-tui)](https://goreportcard.com/report/github.com/audstanley/libgen-tui)

## To build the application, install GoLang **or** Download the binary from the [GitHub Releases](https://github.com/audstanley/libgen-tui/releases)

### Build Process in bash (for Windows/Mac/Linux)

If you are building on Linux, You'll need to install GoLang 

You can use Windows Subsystem for Linux to build - or you can install GoLang natively on Windows, but the native Windows build process is not covered in this README. Hint: Most of what you need is that in the build file. We suggest using [WSL](https://docs.microsoft.com/en-us/windows/wsl/install) if you are building on Windows (then you can just follow the Linux section).

### Linux/Mac/WSL build:

```bash
# sudo apt install git -y # if you don't have git already installed
git clone git@github.com:audstanley/libgen-tui.git
cd libgen-tui
./build.sh
```

### Linux

Once you have the binary on your system, you can cp the binary to the /usr/local/bin directory

```bash
sudo cp build/libgen-amd64-linux /usr/local/bin/libgen
```

Then you can run the application.

```bash
libgen
```

![libgen-tui.gif](libgen-tui.gif)

As you browse through books, The download directory is **.books** in your home directory.  This can be changed in libgen.go, and rebuild the binary to what ever directory you prefer. 
To view the books downloaded:


```bash
cd ~/.books
ls -la
# total 104
# drwxrwxr-x  2 audstanley audstanley  4096 Nov 24 12:45  .
# drwxr-xr-x 47 audstanley audstanley  4096 Nov 24 13:05  ..
# -rw-rw-r--  1 audstanley audstanley 97245 Nov 24 12:45 'Wells, H G - The Time Machine.epub'

```

Then if you have [epy](https://github.com/wustho/epy) installed you can read your books

```bash
epy Wells,\ H\ G\ -\ The\ Time\ Machine.epub
```

We recommend installing [epy](https://github.com/wustho/epy) to read your books in the cli