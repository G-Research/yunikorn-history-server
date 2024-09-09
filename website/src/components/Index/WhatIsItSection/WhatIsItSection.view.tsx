import React from 'react';
import clsx from 'clsx';
import styles from './WhatIsItSection.module.css';
import {FeatureItem} from "@site/src/core/types";

const featuresList: FeatureItem[] = [
    {
        title: 'Historical Data Insights',
        Svg: require('@site/static/img/features/data_trends.svg').default,
        description: (
            <>
                YHS provides detailed insights into past scheduling decisions and resource utilization within Kubernetes
                clusters. This helps teams optimize cluster performance and troubleshoot issues effectively.
            </>
        ),
    },
    {
        title: 'Seamless Integration',
        Svg: require('@site/static/img/features/certification.svg').default,
        description: (
            <>
                YHS is designed to seamlessly integrate with the
                <a href="https://yunikorn.apache.org/" target="_blank"> Yunikorn Scheduler</a>, ensuring smooth data
                collection and storage without disrupting existing workflows or adding significant overhead.
            </>
        ),
    },
    {
        title: 'Open Source Community',
        Svg: require('@site/static/img/features/community.svg').default,
        description: (
            <>
                YHS encourages contributions and collaboration from the community, fostering innovation and
                continuous improvement through shared knowledge and feedback.
            </>
        ),
    },
];


export default function WhatIsItSectionView(): React.JSX.Element {
    return (
        <section>
            <div className={clsx("container padding-bottom--xl", styles.section)}>
                <h2 id="what-is-it" className={clsx("text--center section__ref")}>What’s it?</h2>
                <p className={clsx("text--center padding-horiz--xl", styles.description)}>
                    Yunikorn History Server (YHS) is an ancillary service for K8s Clusters using the Yunikorn
                    Scheduler to persist the state
                    of a Yunikorn-managed cluster to a database, allowing for long-term access to historical data of the cluster's
                    operations (e.g. to view past Applications, resource usage, etc.).
                </p>
                <div className={clsx("row padding-top--md", styles.features)}>
                    {featuresList.map((item, idx) => (
                        <div key={idx} className={clsx('col col--4')}>
                            <div className="card">
                                <div className="card__image text--center padding-top--lg">
                                    <item.Svg className={styles.featureSvg} role="img"/>
                                </div>
                                <div className="card__body text--center">
                                    <h3>{item.title}</h3>
                                    <p>{item.description}</p>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}
