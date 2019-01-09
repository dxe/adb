var path = require('path')
var webpack = require('webpack')
var VueLoaderPlugin = require('vue-loader/lib/plugin')

module.exports = {
  mode: 'development',

  // List of bundles to create. If you want to add a new page, you'll
  // need to also add it here.
  entry: {
    activist_list: './frontend/activist_list',
    event_list: './frontend/event_list',
    event_new: './frontend/event_new',
    flash_message: './frontend/flash_message',
    working_group_list: './frontend/working_group_list',
    circles_list: './frontend/circles_list',
    user_list: './frontend/user_list',
  },

  output: {
    path: path.resolve(__dirname, './dist'),
    filename: '[name].js',
    library: '[name]',
    libraryTarget: 'var',
  },

  plugins: [
    new VueLoaderPlugin(),
  ],

  module: {
    rules: [
      {
        test: /\.vue$/,
        loader: 'vue-loader',
      },
      {
        test: /\.[jt]sx?$/,
        loader: 'ts-loader',
        exclude: /node_modules/,
        options: {
          appendTsSuffixTo: [/\.vue$/],
          reportFiles: ['!frontend/external/**/*.ts'],
        },
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
    extensions: ['.js', '.ts'],
    alias: {
      'vue$': 'vue/dist/vue.esm.js'
    },
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
  module.exports.mode = 'production';
}
