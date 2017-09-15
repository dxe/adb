var path = require('path')
var webpack = require('webpack')

module.exports = {

  // List of bundles to create. If you want to add a new page, you'll
  // need to also add it here.
  entry: {
    activist_list: 'activist_list',
    event_list: 'event_list',
    event_new: 'event_new',
    leaderboard: 'leaderboard',
    flash_message: 'flash_message',
    user_list: 'user_list',
  },

  output: {
    path: path.resolve(__dirname, './dist'),
    filename: '[name].js',
    library: '[name]',
    libraryTarget: 'var',
  },

  module: {
    rules: [
      {
        test: /\.vue$/,
        loader: 'vue-loader',
      },
      {
        test: /\.js$/,
        loader: 'babel-loader',
        include: [
          path.resolve('frontend'),
          path.resolve('node_modules/vue-js-modal'),
        ],
      },
      {
        test: /\.(png|jpg|gif|svg)$/,
        loader: 'file-loader',
        options: {
          name: '[name].[ext]?[hash]',
          publicPath: 'dist/',
        },
      },
      {
        test: /\.css$/,
        use: [
          { loader: "style-loader" },
          { loader: "css-loader" }
        ]
      }
    ]
  },
  resolve: {
    alias: {
      'vue$': 'vue/dist/vue.esm.js'
    },
    modules: ['frontend', 'node_modules'],
  },
  devServer: {
    historyApiFallback: true,
    noInfo: true
  },
  performance: {
    hints: false
  },
  devtool: '#source-map'
}

if (process.env.NODE_ENV === 'production') {
  module.exports.devtool = '#source-map'
  // http://vue-loader.vuejs.org/en/workflow/production.html
  module.exports.plugins = (module.exports.plugins || []).concat([
    new webpack.DefinePlugin({
      'process.env': {
        NODE_ENV: '"production"'
      }
    }),
    new webpack.optimize.UglifyJsPlugin({
      sourceMap: true,
      compress: {
        warnings: false
      }
    }),
    new webpack.LoaderOptionsPlugin({
      minimize: true
    })
  ])
}
