const ModuleFederationPlugin = require("webpack/lib/container/ModuleFederationPlugin");
const mf = require("@angular-architects/module-federation/webpack");
const path = require("path");
const share = mf.share;

const sharedMappings = new mf.SharedMappings();
sharedMappings.register(
  path.join(__dirname, 'tsconfig.json'),
  [/* mapped paths to share */]);

module.exports = {
  output: {
    uniqueName: "yhsComponents",
    publicPath: "auto"
  },
  optimization: {
    runtimeChunk: false
  },   
  resolve: {
    alias: {
      ...sharedMappings.getAliases(),
    }
  },
  experiments: {
    outputModule: true
  },
  plugins: [
    new ModuleFederationPlugin({
        library: { type: "module" },

        // For remotes (please adjust)
        name: "yhsComponents",
        filename: "remoteEntry.js",
        exposes: {
            './AllocationsDrawerComponent': './/src/app/allocations-drawer/allocations-drawer.component.ts',
            './AllocationsDrawerModule': './/src/app/allocations-drawer/allocations-drawer.module.ts',
            './SchedulerService': './/src/app/services/scheduler/scheduler.service.ts'
        },

        shared: share({
          "@angular/core": { singleton: true, strictVersion: true, requiredVersion: 'auto' }, 
          "@angular/common": { singleton: true, strictVersion: true, requiredVersion: 'auto' }, 
          "@angular/common/http": { singleton: true, strictVersion: true, requiredVersion: 'auto' }, 
          "@angular/router": { singleton: true, strictVersion: true, requiredVersion: 'auto' },
          "@angular/material/paginator": {singleton:true, strictVersion: true, requiredVersion: 'auto'},
          "@angular/material/sidenav": {singleton:true, strictVersion: true, requiredVersion: 'auto'},
          "@angular/material/sort":  {singleton:true, strictVersion: true, requiredVersion: 'auto'},
          "@angular/material/table": {singleton:true, strictVersion: true, requiredVersion: 'auto'},
          "@angular/material/tooltip": {singleton:true, strictVersion: true, requiredVersion: 'auto'},
          "@angular/platform-browser": {singleton:true, strictVersion: true, requiredVersion: 'auto'},
          "@angular/material/select": {singleton:true, strictVersion: true, requiredVersion: 'auto'},
          "@angular/material/form-field": {singleton:true, strictVersion: true, requiredVersion: 'auto'},
          "@angular/material/core": {singleton:true, strictVersion: true, requiredVersion: 'auto'},          

          ...sharedMappings.getDescriptors(),
        })
        
    }),
    sharedMappings.getPlugin()
  ],
};
