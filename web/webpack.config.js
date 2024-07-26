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
            './HelloWorldModule': './/src/app/yhs-hello-world/yhs-hello-world.module.ts',
            './TestModule': './/src/app/test/test.module.ts',
            './TestComponent': './/src/app/test/test.component.ts',
            './AllocationsDrawerComponent': './/src/app/allocations-drawer-with-logs/allocations-drawer-with-logs.component.ts',
            './AllocationsDrawerModule': './/src/app/allocations-drawer-with-logs/allocations-drawer-with-logs.module.ts',
        },

        shared: share({
          "@angular/core": { singleton: true, strictVersion: true, requiredVersion: 'auto' }, 
          "@angular/common": { singleton: true, strictVersion: true, requiredVersion: 'auto' }, 
          "@angular/common/http": { singleton: true, strictVersion: true, requiredVersion: 'auto' }, 
          "@angular/router": { singleton: true, strictVersion: true, requiredVersion: 'auto' },

          ...sharedMappings.getDescriptors(),
        })
        
    }),
    sharedMappings.getPlugin()
  ],
};
