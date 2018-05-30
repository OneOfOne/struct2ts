usage: main [<flags>] [<pkg.struct>...]

Flags:
  -h, --help                  Show context-sensitive help (also try --help-long
                              and --help-man).
      --indent="\t"           Output indentation.
  -m, --mark-optional-fields  Add `?` to fields with omitempty.
  -6, --es6                   generate es6 code
  -C, --no-ctor               Don't generate a ctor.
  -T, --no-toObject           Don't generate a Class.toObject() method.
  -E, --no-exports            Don't automatically export the generated types.
  -D, --no-date               Don't automatically handle time.Unix () <-> JS
                              Date().
  -H, --no-helpers            Don't output the helpers.
  -N, --no-default-values     Don't assign default/zero values in the ctor.
  -i, --interface             Only generate an interface (disables all the other
                              options).
  -s, --src-only              Only output the Go code (helpful if you want to
                              edit it yourself).
  -p, --package-name="main"   the package name to use if --src-only is set.
  -k, --keep-temp             Keep the generated Go file, ignored if --src-only
                              is set.
  -o, --out="-"               Write the output to a file instead of stdout.
  -V, --version               Show application version.

Args:
  [<pkg.struct>]  List of structs to convert (github.com/you/auth/users.User,
                  users.User or users.User:AliasUser).

