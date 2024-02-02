# Test Package Files with LXD

Used to test the RPM and DEB packages using LXD across a variety of
distributions.

The image will add the InfluxData repo, install Telegraf, and ensure the
service is running. At that point the new package will get installed and
ensure the service is still running.

Any issues or errors will cause the test to fail.

## CLI

To test an RPM or DEB with a specific image:

```sh
./package-test-lxd --package telegraf_1.21.4-1_amd64.deb --image debian/bullseye
```

To test an RPM or a DEB with a whole set of images:

```sh
./package-test-lxd --package telegraf_1.21.4-1_amd64.deb
```
