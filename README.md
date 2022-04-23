# crypt

**This fork implements experimental support for [age](https://github.com/FiloSottile/age) as an alternative encryption mechanism.**

PGP is still supported. Encryption engine can be selected with `-encryption-engine` flag, with possible values of `pgp` 
and `age`. Note that `-keyring` and `-secret-keyring` file format must correspond to the selected encryption engine (PGP
keyrings for PGP, [recipient files](https://github.com/FiloSottile/age#recipient-files) for age.).

The reason behind that implementation was cumbersome handling of PGP keyrings with `gpg`, 
[lack of maintainers/deprecation of golang.org/x/crypto/openpgp](https://github.com/golang/go/issues/44226)
and general [problems with PGP](https://latacora.micro.blog/2019/07/16/the-pgp-problem.html). 
Notably, some standalone tools exist for the purpose of PGP keys management, like [gpg-tui](https://github.com/orhun/gpg-tui).

For more information about `age`, keys management etc., check https://github.com/FiloSottile/age.

You can use crypt as a command line tool or as a configuration library:

* [crypt cli](bin/crypt)
* [crypt/config](config)

## Demo

Watch Kelsey explain `crypt` in this quick 5 minute video:

[![Crypt Demonstration Video](https://img.youtube.com/vi/zYpqqfuGwW8/0.jpg)](https://www.youtube.com/watch?v=zYpqqfuGwW8)

## Generating gpg keys and keyrings

The crypt cli and config package require gpg keyrings. 

### Create a key and keyring from a batch file

```bash
vim app.batch
```

```
%echo Generating a configuration OpenPGP key
Key-Type: default
Subkey-Type: default
Name-Real: app
Name-Comment: app configuration key
Name-Email: app@example.com
Expire-Date: 0
%pubring .pubring.gpg
%secring .secring.gpg
%commit
%echo done
```

Run the following command:

```bash
gpg2 --batch --armor --gen-key app.batch
```

You should now have two keyrings, `.pubring.gpg` which contains the public keys, and `.secring.gpg` which contains the private keys.

> Note the private key is not protected by a passphrase.
