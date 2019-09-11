const merge = require('webpack-merge')
const UglifyJsPlugin = require('uglifyjs-webpack-plugin')
// const TerserPlugin = require('terser-webpack-plugin'); //TODO Unexpected token: keyword «const
const webpackBaseConfig = require('./webpack.base.config.js')

module.exports = merge(webpackBaseConfig, {
  optimization: {
    minimizer: [
      new UglifyJsPlugin()
    ]
  }
})
