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
        image: 'images/project-social-preview.jpg', // project's social card
        navbar: {
            title: ``,
            logo: {
                alt: 'Project logo',
                src: 'logo/project-logo-light.svg',
                srcDark: 'logo/project-logo-dark.svg',
                width: 140,
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
                            label: 'What’s it?',
                            to: '/#what-is-it',
                        },
                        {
                            label: `Architecture`,
                            to: '/#architecture',
                        },
                        {
                            label: `Community`,
                            to: '/#community',
                        },
                    ],
                },
                {
                    title: 'Contribute',
                    items: [
                        {
                            label: `Start a Discussion`,
                            href: 'https://github.com/G-Research/yunikorn-history-server/discussions'
                        },
                        {
                            label: `Report an Issue`,
                            href: 'https://github.com/G-Research/yunikorn-history-server/issues',
                        },
                        {
                            label: `Create a Pull Request`,
                            href: 'https://github.com/G-Research/yunikorn-history-server/pulls',
                        },
                    ],
                },
                {
                    title: 'More',
                    items: [
                        {
                            label: 'GitHub',
                            href: 'https://github.com/G-Research/yunikorn-history-server',
                        },
                        {
                            label: 'Twitter',
                            href: 'https://twitter.com/oss_gr',
                        },
                        {
                            label: 'G-Research Open-Source',
                            href: 'https://opensource.gresearch.com/',
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
            id: `announcement-bar__support-us`,
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
                content: 'Yunikorn, YHS, Yunikorn History Server, K8S, Kubernetes, Scheduler, History Server',
            },
        ],
    } satisfies Preset.ThemeConfig,
};

export default config;
