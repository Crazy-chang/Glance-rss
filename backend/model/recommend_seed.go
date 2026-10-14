package model

// defaultRecommendSources 内置推荐源（共 30 个）
// HotRank > 0 进入发现页"热门订阅"横滑区，数字越小越靠前
var defaultRecommendSources = []RecommendSource{
	// —— 科技资讯 ——
	{Title: "36氪", FeedURL: "https://www.36kr.com/feed", SiteURL: "https://36kr.com", Category: "科技资讯", Description: "提供科技与商业领域的资讯、观点和深度报道", HotRank: 1},
	{Title: "爱范儿", FeedURL: "https://www.ifanr.com/feed", SiteURL: "https://www.ifanr.com", Category: "科技资讯", Description: "让未来触手可及，关注消费电子与互联网", HotRank: 2},
	{Title: "少数派", FeedURL: "https://sspai.com/feed", SiteURL: "https://sspai.com", Category: "科技资讯", Description: "高品质数字消费指南，效率工具与工作流", HotRank: 3},
	{Title: "极客公园", FeedURL: "https://www.geekpark.net/rss", SiteURL: "https://www.geekpark.net", Category: "科技资讯", Description: "聚焦科技产品与创新公司的深度报道"},
	{Title: "Solidot", FeedURL: "https://www.solidot.org/index.rss", SiteURL: "https://www.solidot.org", Category: "科技资讯", Description: "奇客资讯，报道科技和互联网相关新闻", HotRank: 10},
	{Title: "InfoQ 中文站", FeedURL: "https://www.infoq.cn/feed", SiteURL: "https://www.infoq.cn", Category: "科技资讯", Description: "面向架构师和技术决策者的技术内容平台", HotRank: 6},
	{Title: "IT之家", FeedURL: "https://www.ithome.com/rss/", SiteURL: "https://www.ithome.com", Category: "科技资讯", Description: "科技新闻与数码产品资讯"},
	{Title: "百度实时热点", FeedURL: "https://rss.aishort.top/?type=baidu", SiteURL: "https://top.baidu.com", Category: "科技资讯", Description: "百度搜索实时热点榜单"},

	// —— 技术博客 ——
	{Title: "阮一峰的网络日志", FeedURL: "https://www.ruanyifeng.com/blog/atom.xml", SiteURL: "https://www.ruanyifeng.com/blog/", Category: "技术博客", Description: "科技爱好者周刊与 Web 技术文章", HotRank: 4},
	{Title: "酷壳 CoolShell", FeedURL: "https://coolshell.cn/feed", SiteURL: "https://coolshell.cn", Category: "技术博客", Description: "陈皓的博客，分享编程与技术思考", HotRank: 7},
	{Title: "美团技术团队", FeedURL: "https://tech.meituan.com/feed/", SiteURL: "https://tech.meituan.com", Category: "技术博客", Description: "美团一线工程实践与技术总结"},
	{Title: "腾讯 AlloyTeam", FeedURL: "http://www.alloyteam.com/feed/", SiteURL: "http://www.alloyteam.com", Category: "技术博客", Description: "腾讯 AlloyTeam 前端团队技术博客"},
	{Title: "张鑫旭-鑫空间", FeedURL: "https://www.zhangxinxu.com/wordpress/feed/", SiteURL: "https://www.zhangxinxu.com", Category: "技术博客", Description: "Web 前端开发，CSS/HTML/JavaScript 心得"},
	{Title: "云风的 BLOG", FeedURL: "https://blog.codingnow.com/atom.xml", SiteURL: "https://blog.codingnow.com", Category: "技术博客", Description: "游戏与底层技术、编程实践随笔"},
	{Title: "月光博客", FeedURL: "https://www.williamlong.info/atom.xml", SiteURL: "https://www.williamlong.info", Category: "技术博客", Description: "关注互联网、搜索引擎和信息技术"},
	{Title: "Thoughtworks 洞见", FeedURL: "https://insights.thoughtworks.cn/feed/", SiteURL: "https://insights.thoughtworks.cn", Category: "技术博客", Description: "软件架构、工程效能与数字化转型洞见"},
	{Title: "全栈应用开发:精益实践", FeedURL: "https://www.phodal.com/blog/feeds/rss/", SiteURL: "https://www.phodal.com", Category: "技术博客", Description: "Phodal 的全栈开发与架构实践博客"},

	// —— 数码生活 ——
	{Title: "数字尾巴", FeedURL: "https://www.dgtle.com/feed", SiteURL: "https://www.dgtle.com", Category: "数码生活", Description: "分享美好数字生活，数码评测与体验"},
	{Title: "什么值得买", FeedURL: "https://feed.smzdm.com/", SiteURL: "https://www.smzdm.com", Category: "数码生活", Description: "消费决策门户，好价与商品评测"},

	// —— 财经商业 ——
	{Title: "华尔街见闻", FeedURL: "https://dedicated.wallstreetcn.com/rss.xml", SiteURL: "https://wallstreetcn.com", Category: "财经商业", Description: "全球宏观与市场资讯，7x24 财经快讯", HotRank: 8},
	{Title: "虎嗅网", FeedURL: "https://www.huxiu.com/rss/0.xml", SiteURL: "https://www.huxiu.com", Category: "财经商业", Description: "商业资讯与观点平台"},
	{Title: "金十数据", FeedURL: "https://rsshub.dicomp.net/jin10", SiteURL: "https://www.jin10.com", Category: "财经商业", Description: "全球财经快讯与数据日历"},
	{Title: "财联社", FeedURL: "https://rsshub.dicomp.net/cls/depth/1000", SiteURL: "https://www.cls.cn", Category: "财经商业", Description: "7x24 财经快讯与深度报道"},
	{Title: "第一财经 - 新闻", FeedURL: "https://rsshub.dicomp.net/yicai/news", SiteURL: "https://www.yicai.com", Category: "财经商业", Description: "财经新闻与商业报道"},
	{Title: "7*24小时全球财经", FeedURL: "https://rsshub.dicomp.net/10jqka/realtimenews", SiteURL: "https://www.10jqka.com.cn", Category: "财经商业", Description: "同花顺全球财经实时快讯"},

	// —— 国际新闻 ——
	{Title: "BBC 中文", FeedURL: "https://feeds.bbci.co.uk/zhongwen/simp/rss.xml", SiteURL: "https://www.bbc.com/zhongwen/simp", Category: "国际新闻", Description: "BBC 中文网国际与中国相关报道"},
	{Title: "纽约时报中文网", FeedURL: "https://cn.nytimes.com/rss.html", SiteURL: "https://cn.nytimes.com", Category: "国际新闻", Description: "纽约时报中文国际新闻与深度分析"},
	{Title: "德国之声中文网", FeedURL: "https://rss.dw.com/rdf/rss-chi-all", SiteURL: "https://www.dw.com/zh", Category: "国际新闻", Description: "德国之声中文国际资讯与专题报道"},

	// —— 国内新闻 ——
	{Title: "头条-求是网", FeedURL: "https://plink.anyfeeder.com/qstheory", SiteURL: "http://www.qstheory.cn", Category: "国内新闻", Description: "求是网头条与理论文章"},
	{Title: "新华社新闻_新华网", FeedURL: "https://plink.anyfeeder.com/newscn/whxw", SiteURL: "http://www.news.cn", Category: "国内新闻", Description: "新华社即时新闻报道"},
}
