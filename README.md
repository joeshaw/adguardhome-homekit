# adguardhome-homekit

HomeKit support for [AdGuard Home](https://adguard.com/en/adguard-home/overview.html).

This service publishes a HomeKit switch accessory for enabling or
disabling adblocking.

Once paired with your iOS or Mac home app, you can control it with any
service that integrates with HomeKit, including Siri, Shortcuts, and the
Apple Watch.

## Installing

The tool can be installed with:

    go install github.com/joeshaw/adguardhome-homekit@latest

You will need to create a `config.json` file with your AdGuard Home URL,
username, and password:

```json
{
   "url": "http://192.168.1.100",
   "username": "example",
   "password": "hunter2"
}
```

Then run the service:

    adguardhome-homekit -config config.json

The service will make an initial call to AdGuard Home and update its
state every 15 seconds.

To pair, open up your Home iOS app, click the + icon, choose "Add
Accessory" and then tap "Don't have a Code or Can't Scan?"  You should
see the "switch" under "Nearby Accessories."  Tap that and enter the PIN
00102003 (or whatever you chose in your config file).

## Contributing

Issues and pull requests are welcome.  When filing a PR, please make
sure the code has been run through `gofmt`.

## License

Copyright 2021 Joe Shaw

`adguardhome-homekit` is licensed under the MIT License.  See the
LICENSE file for details.


