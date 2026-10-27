export default defineAppConfig({
  pages: [
    'pages/index/index',
    'pages/discover/index',
    'pages/source/index',
    'pages/article-detail/index',
    'pages/source-article/index',
    'pages/setting/index'
  ],
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: '#FFFFFF',
    navigationBarTitleText: 'Glance',
    navigationBarTextStyle: 'black',
    backgroundColor: '#F7F8FA'
  },
  tabBar: {
    color: '#999999',
    selectedColor: '#5B9DF9',
    backgroundColor: '#FFFFFF',
    borderStyle: 'white',
    list: [
      {
        pagePath: 'pages/index/index',
        text: '文章',
        iconPath: 'assets/tab/article.png',
        selectedIconPath: 'assets/tab/article-active.png'
      },
      {
        pagePath: 'pages/source/index',
        text: '订阅',
        iconPath: 'assets/tab/source.png',
        selectedIconPath: 'assets/tab/source-active.png'
      },
      {
        pagePath: 'pages/setting/index',
        text: '设置',
        iconPath: 'assets/tab/setting.png',
        selectedIconPath: 'assets/tab/setting-active.png'
      }
    ]
  }
})
