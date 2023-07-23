# psitop

psitop - top for [/proc/pressure](https://docs.kernel.org/accounting/psi.html).

Allows you to see resource contention for CPU, IO and memory separately, with
high-resolution 10 second load averages.

![screenshot of psitop](https://github.com/jamespwilliams/psitop/blob/main/_assets/screenshot.png?raw=true)

## Running psitop

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
