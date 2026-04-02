import React, { useState } from "react";
import Layout from "@theme/Layout";
import styles from "./playbooks.module.css";

const playbooks = [
  {
    title: "Kafka to PostgreSQL",
    description:
      "Set up a flowing pipeline from Kafka to PostgreSQL with Avro schema decoding.",
    href: "/playbooks/kafka-to-postgresql",
    tags: ["kafka", "postgresql", "avro", "flowing"],
  },
  {
    title: "HTTP Webhooks",
    description:
      "Accept webhook data over HTTP and store it in a database with validation.",
    href: "/playbooks/http-webhooks",
    tags: ["http", "webhooks", "postgresql", "validation"],
  },
  {
    title: "Build a Weather Tool for AI Assistants",
    description:
      "Build a weather tool that AI assistants can call via Model Context Protocol.",
    href: "/playbooks/mcp-weather-tool",
    tags: ["mcp", "ai", "tools", "weather"],
  },
  {
    title: "Customer Support Agent",
    description:
      "Build an AI support agent with three MCP tools that look up orders, find customer history, and draft replies.",
    href: "/playbooks/customer-support-agent",
    tags: ["mcp", "ai", "sqlite", "sql", "agent"],
  },
];

export default function Playbooks() {
  const [query, setQuery] = useState("");

  const filtered = playbooks.filter((pb) => {
    if (!query) return true;
    const q = query.toLowerCase();
    return (
      pb.title.toLowerCase().includes(q) ||
      pb.description.toLowerCase().includes(q) ||
      pb.tags.some((t) => t.includes(q))
    );
  });

  return (
    <Layout
      title="Playbooks"
      description="Step-by-step walkthroughs for common Qaynaq use cases."
    >
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>Playbooks</h1>
          <p className={styles.subtitle}>
            Step-by-step walkthroughs for common Qaynaq use cases.
          </p>
          <div className={styles.searchWrapper}>
            <svg
              className={styles.searchIcon}
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <input
              type="text"
              className={styles.searchInput}
              placeholder="Search playbooks..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>
        </div>
        <div className={styles.grid}>
          {filtered.map((pb) => (
            <a key={pb.href} href={pb.href} className={styles.card}>
              <h2 className={styles.cardTitle}>{pb.title}</h2>
              <p className={styles.cardDescription}>{pb.description}</p>
              <div className={styles.tags}>
                {pb.tags.map((tag) => (
                  <span key={tag} className={styles.tag}>
                    {tag}
                  </span>
                ))}
              </div>
            </a>
          ))}
          {filtered.length === 0 && (
            <p className={styles.noResults}>No playbooks match your search.</p>
          )}
        </div>
      </div>
    </Layout>
  );
}
