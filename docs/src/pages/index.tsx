import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <h1 className="hero__title">{siteConfig.title}</h1>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/introduction">
            Get Started - 5min ⏱️
          </Link>
        </div>
      </div>
    </header>
  );
}

const FeatureList = [
  {
    title: 'Premium Minimalist',
    description: (
      <>
        Kairo strips away the noise. No borders, no clutter - just structured whitespace 
         and refined typography to keep you in your flow.
      </>
    ),
  },
  {
    title: 'Keyboard First',
    description: (
      <>
        Genuinely fast fuzzy search, Vim bindings, and a powerful command palette. 
        You never have to touch your mouse.
      </>
    ),
  },
  {
    title: 'Extensible Core',
    description: (
      <>
        Lua plugins, a headless CLI API, and a built-in MCP server. 
        Kairo fits perfectly into your existing development workflow.
      </>
    ),
  },
];

function Feature({title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title} Docs`}
      description="Documentation for Kairo - The terminal task manager for developers.">
      <HomepageHeader />
      <main>
        <section className={styles.features}>
          <div className="container">
            <div className="row">
              {FeatureList.map((props, idx) => (
                <Feature key={idx} {...props} />
              ))}
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
