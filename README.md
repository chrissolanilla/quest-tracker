# installation
you gotta go to your asana apps and get your pat and client secrete and id and enable scopes and management and oauth stuff
  ```cd docker
  docker compose up

  cd ../server
  #put in some env vars
  ENV TEMPLATE: 
    ASANA_PROJECT_ID=
    DATABASE_URL= whatever is in your docker compose liek postgres://quest:questpw@localhost:5432/questboard?sslmode=disable
    API_PORT=8080 or whatever is in ur dockef ocmpse
    CORS_ORIGIN=http://localhost:5173
    ASANA_REDIRECT_URI= or whatever you want to put in asana dashboard http://localhost:5173/api/auth/asana/callback 
    
    ASANA_PERSONAL_A_TOKEN=your pat
    ASANA_CLIENT_ID=your client id
    ASANA_CLIENT_SECRET=your client secrete
    POST_LOGIN_REDIRECT=http://localhost:5173/profile or your actual redirect /profile
    ASANA_SCOPES=users:read

  go run .

  cd ../frontend
  npm i
  npm run dev```

