import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
    title: 'Yunikorn History Server (YHS)',
    tagline: 'A service to store and provide historical data for K8S clusters using the Yunikorn scheduler',
    favicon: 'favicon.ico',

    onBrokenLinks: 'throw',
    onBrokenMarkdownLinks: 'warn',

    // Set the production url of your site here
    url: 'https://g-research.github.io',
    // Set the /<baseUrl>/ pathname under which your site is served
    // For GitHub pages deployment, it is often '/<projectName>/'
    baseUrl: '/yunikorn-history-server/',

    // GitHub pages deployment config.
    // If you aren't using GitHub pages, you don't need these.
    organizationName: 'G-Research', // Usually your GitHub org name.
    projectName: 'yunikorn-history-server', // Usually your repo name.

    i18n: {
        defaultLocale: 'en',
        locales: ['en'],
    },

    presets: [
        [
            'classic',
            {
                docs: false,
                theme: {
                    customCss: [
                        './src/css/theming.css',
                        './src/css/global.css',
                        './src/css/announcement-bar.css',
                    ],
                },
            } satisfies Preset.Options,
        ],
    ],

    themeConfig: {
        image: 'images/social-preview.jpg',
        navbar: {
            title: ``,
            logo: {
                alt: 'Yunikorn History Server',
                src: 'navbar/logo-light-bg.svg',
                srcDark: 'navbar/logo-dark-bg.svg',
                width: 112,
            },
            items: [
                // left
                {
                    to: '/#what-is-it',
                    label: 'What’s it?',
                    position: 'left',
                    activeBaseRegex: `dummy-never-match`,
                },
                {
                    to: '/#architecture',
                    label: 'Architecture',
                    position: 'left',
                    activeBaseRegex: `dummy-never-match`,
                },
                {
                    to: '/#community',
                    label: 'Community',
                    position: 'left',
                    activeBaseRegex: `dummy-never-match`,
                },
                // right
                {
                    href: 'https://github.com/G-Research/yunikorn-history-server',
                    label: 'GitHub',
                    position: 'right',
                },
                {
                    href: 'https://twitter.com/oss_gr',
                    label: 'Twitter',
                    position: 'right',
                },
            ],
        },
        footer: {
            links: [
                {
                    title: 'About',
                    items: [
                        {
                            to: '/#what-is-it',
                            label: 'What’s it?',
                        },
                        {
                            to: '/#architecture',
                            label: `Architecture`,
                        },
                        {
                            to: '/#community',
                            label: `Community`,
                        },
                    ],
                },
                {
                    title: 'Contribute',
                    items: [
                        {
                            href: 'https://github.com/G-Research/yunikorn-history-server/discussions',
                            label: `Start a Discussion`,
                        },
                        {
                            href: 'https://github.com/G-Research/yunikorn-history-server/issues',
                            label: `Report an Issue`,
                        },
                        {
                            href: 'https://github.com/G-Research/yunikorn-history-server/pulls',
                            label: `Create a Pull Request`,
                        },
                    ],
                },
                {
                    title: 'More',
                    items: [
                        {
                            href: 'https://github.com/G-Research/yunikorn-history-server',
                            label: 'GitHub',
                        },
                        {
                            href: 'https://twitter.com/oss_gr',
                            label: 'Twitter',
                        },
                        {
                            href: 'https://opensource.gresearch.com/',
                            label: 'G-Research Open-Source',
                        },
                    ],
                },
            ],
            style: 'light',
            logo: {
                alt: 'G-Research Open-Source Software',
                src: 'https://github.com/G-Research/brand/raw/main/logo/GR-OSS/logo.svg',
                srcDark: 'https://github.com/G-Research/brand/raw/main/logo/GR-OSS/logo-dark-bg.svg',
                href: 'https://opensource.gresearch.com/',
            },
            copyright: `Copyright © ${new Date().getFullYear()} G-Research`,
        },
        announcementBar: {
            // https://docusaurus.io/docs/api/themes/configuration#announcement-bar
            id: `announcement-bar__support-project`,
            content: `Support the <a href="https://github.com/G-Research/yunikorn-history-server" target="_blank">project ⭐️!</a>`,
            isCloseable: true,
        },
        colorMode: {
            defaultMode: 'light',
            disableSwitch: false,
            respectPrefersColorScheme: true,
        },
        prism: {
            theme: prismThemes.github,
            darkTheme: prismThemes.dracula,
            defaultLanguage: 'bash',
            additionalLanguages: ['python', 'powershell'],
        },
        metadata: [
            {
                name: 'twitter:card',
                content: 'summary',
            },
            {
                name: 'keywords',
                content: 'Yunikorn History Server, YHS, Kubernetes, K8s, Apache YuniKorn, Scheduler',
            },
        ],
    } satisfies Preset.ThemeConfig,
};

export default config;
