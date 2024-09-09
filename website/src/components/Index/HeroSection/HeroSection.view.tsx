import React from "react";
import clsx from "clsx";
import ThemedImage from '@theme/ThemedImage';
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Link from "@docusaurus/Link";
import useBaseUrl from "@docusaurus/useBaseUrl";
import {IconBrandGithub, IconHelpSquareRounded} from '@tabler/icons-react';
import styles from './HeroSection.module.css';

export default function HeroSectionView(): React.JSX.Element {
    const {siteConfig} = useDocusaurusContext();
    return (
        <header className={clsx('hero', 'padding-vert--lg', styles.heroBanner)}>
            <div className="container">
                <div className="row row--no-gutters">
                    <div className={clsx("col padding-vert--xl margin-vert--xl", styles.heroHeadingColumn)}>
                        <h1 className="hero__title">{siteConfig.title}</h1>
                        <p className="hero__subtitle">{siteConfig.tagline}</p>
                        <div className={styles.buttons}>
                            <Link
                                className="button button--primary button--lg shadow--md"
                                to="/#what-is-it">
                                <IconHelpSquareRounded size={18} color="currentcolor" stroke={3}/> What's it?
                            </Link>
                            <Link
                                className="button button--secondary button--lg shadow--md"
                                to="https://github.com/G-Research/yunikorn-history-server">
                                <IconBrandGithub size={18} color="currentcolor" stroke={3}/> GitHub
                            </Link>
                        </div>
                    </div>
                    <div className={clsx("col", styles.heroIconColumn)}>
                        <div className={styles.heroIconContainer}>
                            <div className={styles.heroIconBackground}></div>
                            <ThemedImage
                                className={styles.heroIcon} width={256} height={256}
                                alt="YHS icon"
                                sources={{
                                    light: useBaseUrl('/img/icon/project-icon-light-bg-2.svg'),
                                    dark: useBaseUrl('/img/icon/project-icon-dark-bg-2.svg'),
                                }}
                            />
                        </div>
                    </div>
                </div>
            </div>
        </header>
    );
}
