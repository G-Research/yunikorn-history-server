import React from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import useBaseUrl from "@docusaurus/useBaseUrl";
import {IconBrandGithub, IconHelpSquareRounded} from '@tabler/icons-react';
import styles from './HeroSection.module.css';


export default function HeroSectionView(): React.JSX.Element {
    return (
        <header>
            <div className={clsx("padding--lg", styles.topBanner)}>
                <h1 className={styles.topBannerTitle}>
                    Yunikorn History Server (YHS)
                </h1>
            </div>
            <div className={clsx("padding-vert--xl padding-horiz--md", styles.hero)}>
                <div className="container">
                    <div className={styles.heroInner}>
                        <img
                            alt='YHS logo'
                            className={clsx("padding--sm", styles.heroLogo)}
                            src={useBaseUrl('/img/logo/project-logo-dark-bg.svg')}
                            width="500"
                            height="225"
                        />
                        <span className={styles.heroTitle}>
                            <b>Store</b> and <b>access</b> historical data of your <b>K8s clusters</b> using <b>Apache Yunikorn</b>
                        </span>
                    </div>
                    <div className={clsx("margin-top--lg", styles.heroCtas)}>
                        <Link
                            className="button button--primary button--lg shadow--md"
                            to="/#what-is-it">
                            <IconHelpSquareRounded size={18} color="currentcolor" stroke={3}/> What is it?
                        </Link>
                        <Link
                            className="button button--secondary button--lg shadow--md"
                            to="https://github.com/G-Research/yunikorn-history-server">
                            <IconBrandGithub size={18} color="currentcolor" stroke={3}/> GitHub
                        </Link>
                        <span className={styles.gitHubButtonWrapper}>
                            <iframe
                                className={styles.gitHubButton}
                                src="https://ghbtns.com/github-btn.html?user=G-Research&amp;repo=yunikorn-history-server&amp;type=star&amp;count=true&amp;size=large"
                                width={160}
                                height={30}
                                title="GitHub Stars"
                            />
                    </span>
                    </div>
                </div>
            </div>
        </header>
    );
}
