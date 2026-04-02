// @ts-check

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Qaynaq",
  tagline: "The fastest way to connect your data to AI",
  favicon: "img/favicon.png",

  url: "https://qaynaq.io",
  baseUrl: "/",

  organizationName: "qaynaq",
  projectName: "qaynaq",

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

  markdown: {
    mermaid: true,
  },

  themes: ["@docusaurus/theme-mermaid"],

  plugins: [
    [
      "@docusaurus/plugin-content-docs",
      {
        id: "playbooks",
        path: "playbooks",
        routeBasePath: "playbooks",
        sidebarPath: "./playbooks-sidebars.js",
        editUrl:
          "https://github.com/qaynaq/qaynaq/tree/main/website/",
      },
    ],
  ],

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: "./sidebars.js",
          editUrl:
            "https://github.com/qaynaq/qaynaq/tree/main/website/",
        },
        theme: {
          customCss: "./src/css/custom.css",
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: "Qaynaq",
        logo: {
          alt: "Qaynaq Logo",
          src: "img/mascot-small.png",
          srcDark: "img/mascot-small.png",
        },
        items: [
          {
            type: "docSidebar",
            sidebarId: "docsSidebar",
            position: "left",
            label: "Docs",
          },
          {
            to: "/playbooks",
            label: "Playbooks",
            position: "left",
          },
          {
            href: "https://github.com/qaynaq/qaynaq/releases/latest",
            position: "right",
            className: "navbar-download-link",
            "aria-label": "Download",
          },
          {
            href: "https://github.com/qaynaq/qaynaq",
            position: "right",
            className: "navbar-github-link",
            "aria-label": "GitHub repository",
          },
        ],
      },
      footer: {
        style: "dark",
        copyright: `Copyright © ${new Date().getFullYear()} Qaynaq. Apache 2.0 License.`,
      },
      prism: {
        theme: require("prism-react-renderer").themes.github,
        darkTheme: require("prism-react-renderer").themes.dracula,
        additionalLanguages: ["bash", "yaml", "sql"],
      },
      colorMode: {
        defaultMode: "light",
        respectPrefersColorScheme: true,
      },
    }),
};

module.exports = config;
