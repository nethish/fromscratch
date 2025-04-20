# UUID

## UUID V1
* Uses time + mac address to generate UUID

## UUID V2
* Instead of some time bits, it uses UID or GID

## UUID V3
* Gives same UUID for given namespace + name
* It's like hashing
* Uses MD5

## UUID V4
* Fully random

## UUID V5
* Same as V3, but uses SHA-1

## OID - Object ID - Mongo
* 24 character hex string - 12 bytes
* Timestamp 4 bytes to make it sortable
* Machine Identifier - 5 bytes
* Process ID - 2 bytes
* Counter - 3 bytes
* Example - 507f191e810c19729de860ea

