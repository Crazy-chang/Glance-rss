// babel-preset-taro 更多选项见 https://docs.taro.zone/docs/next/babel-config
module.exports = {
  presets: [
    [
      'taro',
      {
        framework: 'react',
        ts: true,
        compiler: 'webpack5'
      }
    ]
  ]
}
