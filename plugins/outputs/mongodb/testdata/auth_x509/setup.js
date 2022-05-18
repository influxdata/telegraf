const conn = new Mongo();
const db = conn.getDB("admin");
// createUser normally requires a password unless $external is used
// the CN value was found via: openssl x509 -in client.pem -noout -subject -nameopt RFC2253 | sed 's/subject=//g'
db.getSiblingDB("$external").runCommand({ createUser: "CN=localdomain", roles: [{ role: "root", db: "admin" }] });
