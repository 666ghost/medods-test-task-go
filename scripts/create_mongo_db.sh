echo "running"
mongo <<EOF
   use admin;
   admin = db.getSiblingDB("admin");
   admin.createUser(
     {
	user: "admin",
        pwd: "password",
        roles: [ "root", "userAdminAnyDatabase", "dbAdminAnyDatabase", "readWriteAnyDatabase" ]
     });
     db.getSiblingDB("admin").auth("admin", "password");
EOF
