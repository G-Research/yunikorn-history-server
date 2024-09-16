# YuniKorn Histry Server Components

YuniKorn Histry Server Components provides a remote components that are consumed by the [YuniKorn Web](https://github.com/apache/yunikorn-web/)

## Development environment setup
### Dependencies

- [Node.js](https://nodejs.org/en/)
- [Angular CLI](https://github.com/angular/angular-cli)

For managing node packages you can use `npm`, `yarn` or `pnpm`. Run `npm install` to set up all necessary dependencies.

### Development Server

Run `npm start` for a dev server. Remote components will be served from this path `http://localhost:3100/remoteEntry.js`. The application will automatically reload if you change any of the source files.

### JSON Server

To run a mock server for local development, follow these steps:

**Start the JSON Server**:

   - **Using Makefile**: you can start the server by running:
     ```sh
     make mock-server
     ```

   - **Using npm**: If you are in the `./web` directory, you can run the JSON Server directly with npm by using:
     ```sh
     npm run start:json-server
     ```

   This will start the JSON Server and serve mock data. You can access it at `http://localhost:3000`.

### Build

Run `make web-build` from the project root or `npm run build`. Build output is set to `/assets` folder in project root as it will be served from the YHS server.

## Further help
To get more help on the Angular CLI use `ng help` or go check out the [Angular CLI README](https://github.com/angular/angular-cli/blob/master/README.md).

## Code scaffolding
Run `ng generate component component-name` to generate a new component.

You can also use `ng generate directive|pipe|service|class|guard|interface|enum|module`.
