## Decryption

### AES

Currently, `AES128`, `AES192` and `AES256` are supported. You can (optionally)
specify a _mode_ and _padding_ algorithm in the form`AES256[/mode[/<padding>]]`.

Besides the plain AES (no mode), the following _modes_ are supported:

- CBC block cipher
- CFB stream cipher
- CTR stream cipher
- OFB stream cipher

You can also specify one of the following paddings:

- PKCS#5
- PKCS#7

So for plain AES256 specify `AES256` while for AES128 with CBC and PKCS#5
padding you should use `AES128/CBC/PKCS#5`.