# psitop

psitop - top for [/proc/pressure](https://docs.kernel.org/accounting/psi.html).

Allows you to see resource contention for CPU, IO and memory separately, with
high-resolution 10 second load averages.

![screenshot of psitop](https://github.com/jamespwilliams/psitop/blob/main/_assets/screenshot.png?raw=true)

## Running psitop

First, note that psitop needs to read `/proc/pressure`, which requires Linux
kernel version 4.20 or higher. Your distribution might disable /proc/pressure
by default - if so, you'll need to enable it before using this tool.

### Go

```
go install https://github.com/jamespwilliams/psitop@latest
psitop
```

### Nix

If you have Nix installed and flakes enabled:

```
nix run github:jamespwilliams/psitop
```

## Usage

Use the keybindings shown in the interface.
