About exe-delegate
==================

This program can create a delegate executable to a script / exe file.


Build
======

```
go build
```


Usage
======

It is supported that `foo.exe` is located in `C:\Program Files\foo\bin\foo.exe`. You wanna to use `foo` directly from `cmd`, while you don't like to add `C:\Program Files\foo\bin` to `PATH`. Then you can use `exe-delegate` in this way:

```
exe-delegate -o c:\windows\system32\foo "C:\Program Files\foo\bin\foo.exe"
```

This command will produce a delegate in `c:\windows\system32\foo` for `C:\Program Files\foo\bin\foo.exe`.


