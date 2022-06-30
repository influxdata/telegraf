const conn = new Mongo();
const db = conn.getDB('admin');
db.createUser({ user: 'root', pwd: 'changeme', roles: [{ role: 'root', db: 'admin' }] });
