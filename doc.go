/*
Package partaj is a minimal event stream processing framework for reading, processing and writing events,
using a DSL inspried by functional programming and more cumbersome event processing systems such as Apache Kafka.

Basics

An event processing system consists of event origins, called sources,
connected to event destinations, called sinks, that are connected to each other
and processing steps in stages. A group of sources, stages and sinks form a
contained event processing system, called a pipeline.

*/
package partaj
