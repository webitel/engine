const path = require('path')
const HtmlWebpackPlugin = require('html-webpack-plugin')

module.exports = {
    entry: [ './src/test'],
    output: {
        library: 'webitel',
        // libraryTarget: 'umd',
        filename: 'webitel.js',
        // auxiliaryComment: 'Test Comment'
    },
    devtool: 'source-map',
    devServer: {
        contentBase: path.join(__dirname, 'dist'),
        compress: true,
        host: "0.0.0.0",
        port: 9000
    },
    module: {
        rules: [{
            enforce: 'pre',
            test: /\.js$/,
            loader: 'source-map-loader'
        }, {
            enforce: 'pre',
            test: /\.ts$/,
            exclude: /node_modules/,
            loader: 'tslint-loader'
        }, {
            test: /\.ts$/,
            loader: 'ts-loader'
        }]
    },
    resolve: {
        alias: {
            // jssip: path.resolve(__dirname, '../node_modules/jssip'),
            // sipjs: path.resolve(__dirname, '../node_modules/sip.js/dist/sip.min.js')
        },
        modules: [
            'node_modules',
            path.resolve(__dirname, '../src')
        ],
        extensions: ['.js', '.ts'],
    },
    plugins: [
        new HtmlWebpackPlugin({
            // inject: true,
            template: 'index.html'
        })
    ]
}
