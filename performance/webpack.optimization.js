/**
 * Webpack Optimization Configuration for Taishang Laojun AI Platform
 * 
 * This configuration optimizes the build process for:
 * - Bundle size reduction
 * - Code splitting
 * - Tree shaking
 * - Asset optimization
 * - Regional deployment optimization
 */

const path = require('path');
const webpack = require('webpack');
const TerserPlugin = require('terser-webpack-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const CompressionPlugin = require('compression-webpack-plugin');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;
const { SubresourceIntegrityPlugin } = require('webpack-subresource-integrity');

const optimizationConfig = {
  // Production optimizations
  optimization: {
    minimize: true,
    minimizer: [
      // JavaScript minification
      new TerserPlugin({
        terserOptions: {
          compress: {
            drop_console: process.env.NODE_ENV === 'production',
            drop_debugger: true,
            pure_funcs: ['console.log', 'console.info', 'console.debug'],
            passes: 2
          },
          mangle: {
            safari10: true
          },
          format: {
            comments: false
          }
        },
        extractComments: false,
        parallel: true
      }),
      
      // CSS minification
      new CssMinimizerPlugin({
        minimizerOptions: {
          preset: [
            'default',
            {
              discardComments: { removeAll: true },
              normalizeWhitespace: true,
              colormin: true,
              convertValues: true,
              discardDuplicates: true,
              discardEmpty: true,
              mergeRules: true,
              minifyFontValues: true,
              minifyGradients: true,
              minifyParams: true,
              minifySelectors: true,
              reduceIdents: true,
              reduceInitial: true,
              reduceTransforms: true,
              svgo: true
            }
          ]
        }
      })
    ],
    
    // Code splitting configuration
    splitChunks: {
      chunks: 'all',
      minSize: 20000,
      maxSize: 150000,
      minChunks: 1,
      maxAsyncRequests: 30,
      maxInitialRequests: 30,
      enforceSizeThreshold: 50000,
      cacheGroups: {
        // Vendor libraries
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          chunks: 'all',
          priority: 20,
          reuseExistingChunk: true
        },
        
        // React and related libraries
        react: {
          test: /[\\/]node_modules[\\/](react|react-dom|react-router)[\\/]/,
          name: 'react',
          chunks: 'all',
          priority: 30,
          reuseExistingChunk: true
        },
        
        // UI libraries
        ui: {
          test: /[\\/]node_modules[\\/](@mui|@emotion|styled-components)[\\/]/,
          name: 'ui',
          chunks: 'all',
          priority: 25,
          reuseExistingChunk: true
        },
        
        // Localization libraries
        i18n: {
          test: /[\\/]node_modules[\\/](react-i18next|i18next|@formatjs)[\\/]/,
          name: 'i18n',
          chunks: 'all',
          priority: 25,
          reuseExistingChunk: true
        },
        
        // Common modules
        common: {
          name: 'common',
          minChunks: 2,
          chunks: 'all',
          priority: 10,
          reuseExistingChunk: true,
          enforce: true
        },
        
        // CSS chunks
        styles: {
          name: 'styles',
          type: 'css/mini-extract',
          chunks: 'all',
          priority: 15,
          reuseExistingChunk: true
        }
      }
    },
    
    // Runtime chunk optimization
    runtimeChunk: {
      name: 'runtime'
    },
    
    // Module concatenation (scope hoisting)
    concatenateModules: true,
    
    // Tree shaking
    usedExports: true,
    sideEffects: false,
    
    // Module IDs optimization
    moduleIds: 'deterministic',
    chunkIds: 'deterministic'
  },
  
  // Performance budgets
  performance: {
    maxAssetSize: 250000, // 250KB
    maxEntrypointSize: 250000, // 250KB
    hints: process.env.NODE_ENV === 'production' ? 'error' : 'warning',
    assetFilter: function(assetFilename) {
      return !assetFilename.endsWith('.map');
    }
  },
  
  // Plugins for optimization
  plugins: [
    // Compression plugins
    new CompressionPlugin({
      algorithm: 'gzip',
      test: /\.(js|css|html|svg)$/,
      threshold: 8192,
      minRatio: 0.8,
      deleteOriginalAssets: false
    }),
    
    new CompressionPlugin({
      filename: '[path][base].br',
      algorithm: 'brotliCompress',
      test: /\.(js|css|html|svg)$/,
      compressionOptions: {
        level: 11
      },
      threshold: 8192,
      minRatio: 0.8,
      deleteOriginalAssets: false
    }),
    
    // Bundle analyzer (only in analyze mode)
    ...(process.env.ANALYZE ? [
      new BundleAnalyzerPlugin({
        analyzerMode: 'static',
        openAnalyzer: false,
        reportFilename: 'bundle-report.html'
      })
    ] : []),
    
    // Subresource integrity
    new SubresourceIntegrityPlugin({
      hashFuncNames: ['sha256', 'sha384'],
      enabled: process.env.NODE_ENV === 'production'
    }),
    
    // Define plugin for optimization flags
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV),
      'process.env.REGION': JSON.stringify(process.env.REGION || 'us-east-1'),
      '__DEV__': process.env.NODE_ENV !== 'production',
      '__PROD__': process.env.NODE_ENV === 'production'
    }),
    
    // Module federation for micro-frontends
    new webpack.container.ModuleFederationPlugin({
      name: 'taishanglaojun',
      filename: 'remoteEntry.js',
      exposes: {
        './App': './src/App',
        './Localization': './src/components/Localization',
        './Compliance': './src/components/Compliance'
      },
      shared: {
        react: {
          singleton: true,
          requiredVersion: '^18.0.0'
        },
        'react-dom': {
          singleton: true,
          requiredVersion: '^18.0.0'
        },
        'react-router-dom': {
          singleton: true,
          requiredVersion: '^6.0.0'
        }
      }
    })
  ],
  
  // Resolve optimizations
  resolve: {
    // Module resolution optimization
    modules: [
      path.resolve(__dirname, 'src'),
      'node_modules'
    ],
    
    // Extension resolution order
    extensions: ['.js', '.jsx', '.ts', '.tsx', '.json'],
    
    // Alias for common paths
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@components': path.resolve(__dirname, 'src/components'),
      '@utils': path.resolve(__dirname, 'src/utils'),
      '@services': path.resolve(__dirname, 'src/services'),
      '@hooks': path.resolve(__dirname, 'src/hooks'),
      '@assets': path.resolve(__dirname, 'src/assets'),
      '@locales': path.resolve(__dirname, 'src/locales'),
      '@types': path.resolve(__dirname, 'src/types')
    },
    
    // Fallback for Node.js modules
    fallback: {
      crypto: require.resolve('crypto-browserify'),
      stream: require.resolve('stream-browserify'),
      buffer: require.resolve('buffer')
    }
  },
  
  // Module rules for optimization
  module: {
    rules: [
      // JavaScript/TypeScript optimization
      {
        test: /\.(js|jsx|ts|tsx)$/,
        exclude: /node_modules/,
        use: [
          {
            loader: 'babel-loader',
            options: {
              presets: [
                ['@babel/preset-env', {
                  targets: {
                    browsers: ['> 1%', 'last 2 versions', 'not ie <= 11']
                  },
                  modules: false,
                  useBuiltIns: 'usage',
                  corejs: 3
                }],
                ['@babel/preset-react', {
                  runtime: 'automatic'
                }],
                '@babel/preset-typescript'
              ],
              plugins: [
                '@babel/plugin-proposal-class-properties',
                '@babel/plugin-proposal-object-rest-spread',
                '@babel/plugin-syntax-dynamic-import',
                ['@babel/plugin-transform-runtime', {
                  regenerator: true
                }],
                // Tree shaking for lodash
                ['babel-plugin-lodash', {
                  id: ['lodash', 'recompose']
                }],
                // Import optimization
                ['babel-plugin-import', {
                  libraryName: '@mui/material',
                  libraryDirectory: '',
                  camel2DashComponentName: false
                }, 'core'],
                ['babel-plugin-import', {
                  libraryName: '@mui/icons-material',
                  libraryDirectory: '',
                  camel2DashComponentName: false
                }, 'icons']
              ],
              cacheDirectory: true,
              cacheCompression: false
            }
          }
        ]
      },
      
      // CSS optimization
      {
        test: /\.css$/,
        use: [
          process.env.NODE_ENV === 'production' 
            ? MiniCssExtractPlugin.loader 
            : 'style-loader',
          {
            loader: 'css-loader',
            options: {
              modules: {
                auto: true,
                localIdentName: process.env.NODE_ENV === 'production' 
                  ? '[hash:base64:8]' 
                  : '[name]__[local]--[hash:base64:5]'
              },
              importLoaders: 1
            }
          },
          {
            loader: 'postcss-loader',
            options: {
              postcssOptions: {
                plugins: [
                  'autoprefixer',
                  'cssnano',
                  'postcss-preset-env'
                ]
              }
            }
          }
        ]
      },
      
      // Image optimization
      {
        test: /\.(png|jpe?g|gif|svg|webp|avif)$/i,
        type: 'asset',
        parser: {
          dataUrlCondition: {
            maxSize: 8192 // 8KB
          }
        },
        generator: {
          filename: 'assets/images/[name].[hash:8][ext]'
        },
        use: [
          {
            loader: 'image-webpack-loader',
            options: {
              mozjpeg: {
                progressive: true,
                quality: 85
              },
              optipng: {
                enabled: false
              },
              pngquant: {
                quality: [0.65, 0.90],
                speed: 4
              },
              gifsicle: {
                interlaced: false
              },
              webp: {
                quality: 85
              },
              svgo: {
                plugins: [
                  {
                    name: 'removeViewBox',
                    active: false
                  }
                ]
              }
            }
          }
        ]
      },
      
      // Font optimization
      {
        test: /\.(woff|woff2|eot|ttf|otf)$/i,
        type: 'asset/resource',
        generator: {
          filename: 'assets/fonts/[name].[hash:8][ext]'
        }
      }
    ]
  },
  
  // Cache configuration
  cache: {
    type: 'filesystem',
    buildDependencies: {
      config: [__filename]
    },
    cacheDirectory: path.resolve(__dirname, '.webpack-cache'),
    compression: 'gzip'
  },
  
  // Experiments
  experiments: {
    topLevelAwait: true,
    outputModule: true
  }
};

// Regional optimization configurations
const regionalConfigs = {
  'us-east-1': {
    // US-specific optimizations
    externals: {
      // Use CDN for common libraries in US
      'react': 'React',
      'react-dom': 'ReactDOM'
    }
  },
  
  'eu-central-1': {
    // EU-specific optimizations (GDPR compliance)
    plugins: [
      new webpack.DefinePlugin({
        'process.env.GDPR_ENABLED': 'true'
      })
    ]
  },
  
  'ap-east-1': {
    // Asia-specific optimizations
    optimization: {
      splitChunks: {
        cacheGroups: {
          // Separate chunk for Chinese fonts
          chineseFonts: {
            test: /[\\/]assets[\\/]fonts[\\/]chinese/,
            name: 'chinese-fonts',
            chunks: 'all',
            priority: 40
          }
        }
      }
    }
  }
};

// Merge regional configuration
function getOptimizedConfig(region = 'us-east-1') {
  const regionalConfig = regionalConfigs[region] || {};
  
  return {
    ...optimizationConfig,
    ...regionalConfig,
    optimization: {
      ...optimizationConfig.optimization,
      ...regionalConfig.optimization
    },
    plugins: [
      ...optimizationConfig.plugins,
      ...(regionalConfig.plugins || [])
    ]
  };
}

module.exports = {
  optimizationConfig,
  regionalConfigs,
  getOptimizedConfig
};