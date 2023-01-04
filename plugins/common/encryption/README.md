### Plain text

Set `cipher` to `none` if the secrets are transmitted as plain-text. No further
options are required.

### Advanced Encryption Standard (AES)

Currently the following AES ciphers are supported

- `AES128`: plain AES with 128-bit key length
- `AES128/CBC`: 128-bit key in _CBC_ block mode without padding
- `AES128/CBC/PKCS#5`: 128-bit key in _CBC_ block mode with _PKCS#5_ padding
- `AES128/CBC/PKCS#7`: 128-bit key in _CBC_ block mode with _PKCS#7_ padding
- `AES128/CFB`: 128-bit key in _CFB_ streaming mode without padding
- `AES128/CFB/PKCS#5`: 128-bit key in _CFB_ streaming mode with _PKCS#5_ padding
- `AES128/CFB/PKCS#7`: 128-bit key in _CFB_ streaming mode with _PKCS#7_ padding
- `AES128/CTR`: 128-bit key in _CTR_ streaming mode without padding
- `AES128/CTR/PKCS#5`: 128-bit key in _CTR_ streaming mode with _PKCS#5_ padding
- `AES128/CTR/PKCS#7`: 128-bit key in _CTR_ streaming mode with _PKCS#7_ padding
- `AES128/OFB`: 128-bit key in _OFB_ streaming mode without padding
- `AES128/OFB/PKCS#5`: 128-bit key in _OFB_ streaming mode with _PKCS#5_ padding
- `AES128/OFB/PKCS#7`: 128-bit key in _OFB_ streaming mode with _PKCS#7_ padding
- `AES192`: plain AES with 192-bit key length
- `AES192/CBC`: 192-bit key in _CBC_ block mode without padding
- `AES192/CBC/PKCS#5`: 192-bit key in _CBC_ block mode with _PKCS#5_ padding
- `AES192/CBC/PKCS#7`: 192-bit key in _CBC_ block mode with _PKCS#7_ padding
- `AES192/CFB`: 192-bit key in _CFB_ streaming mode without padding
- `AES192/CFB/PKCS#5`: 192-bit key in _CFB_ streaming mode with _PKCS#5_ padding
- `AES192/CFB/PKCS#7`: 192-bit key in _CFB_ streaming mode with _PKCS#7_ padding
- `AES192/CTR`: 192-bit key in _CTR_ streaming mode without padding
- `AES192/CTR/PKCS#5`: 192-bit key in _CTR_ streaming mode with _PKCS#5_ padding
- `AES192/CTR/PKCS#7`: 192-bit key in _CTR_ streaming mode with _PKCS#7_ padding
- `AES192/OFB`: 192-bit key in _OFB_ streaming mode without padding
- `AES192/OFB/PKCS#5`: 192-bit key in _OFB_ streaming mode with _PKCS#5_ padding
- `AES192/OFB/PKCS#7`: 192-bit key in _OFB_ streaming mode with _PKCS#7_ padding
- `AES256`: plain AES with 256-bit key length
- `AES256/CBC`: 256-bit key in _CBC_ block mode without padding
- `AES256/CBC/PKCS#5`: 256-bit key in _CBC_ block mode with _PKCS#5_ padding
- `AES256/CBC/PKCS#7`: 256-bit key in _CBC_ block mode with _PKCS#7_ padding
- `AES256/CFB`: 256-bit key in _CFB_ streaming mode without padding
- `AES256/CFB/PKCS#5`: 256-bit key in _CFB_ streaming mode with _PKCS#5_ padding
- `AES256/CFB/PKCS#7`: 256-bit key in _CFB_ streaming mode with _PKCS#7_ padding
- `AES256/CTR`: 256-bit key in _CTR_ streaming mode without padding
- `AES256/CTR/PKCS#5`: 256-bit key in _CTR_ streaming mode with _PKCS#5_ padding
- `AES256/CTR/PKCS#7`: 256-bit key in _CTR_ streaming mode with _PKCS#7_ padding
- `AES256/OFB`: 256-bit key in _OFB_ streaming mode without padding
- `AES256/OFB/PKCS#5`: 256-bit key in _OFB_ streaming mode with _PKCS#5_ padding
- `AES256/OFB/PKCS#7`: 256-bit key in _OFB_ streaming mode with _PKCS#7_ padding

Additional to the cipher, you need to provide the encryption `key` and
initialization vector `init_vector` to be able to decrypt the data.
In case you are using password-based key derivation, `key`
(and possibly `init_vector`) can be omitted. Take a look at the
[password-based key derivation section](#password-based-key-derivation).

### Password-based key derivation

Alternatively to providing a `key` (and `init_vector`) the key (and vector)
can be derived from a given password. Currently the following algorithms are
supported for `kdf_algorithm`:

 - `PBKDF2-HMAC-SHA256` for `key` only, no `init_vector` created

You also need to provide the `password` to derive the key from as well as the
`salt` and `iterations` used.
__Please note:__ All parameters must match the encryption side to derive the
same key in Telegraf!