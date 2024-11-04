The Unicorn History Server (UHS) is a standalone service that enhances the capabilities of the
Yunikorn Scheduler by providing long-term persistence of cluster operational data.
It achieves this by listening for events from the Yunikorn Scheduler and persisting them to a database.

![UHS Architecture](https://github.com/G-Research/unicorn-history-server/raw/main/uhs-architecture.png)

UHS is composed of three main components:

1. **Event Collector:** This component is responsible for listening to events stream API from the Yunikorn Scheduler  
   and persisting them to the database.
   It ensures that all significant operations performed by the scheduler
   are recorded for future analysis.

2. **REST API:** This component serves as the interface for retrieving historical data.
   It provides endpoints that return data about past applications, resource usage, and other operational metrics. Also provides
   querying capabilities to filter and retrieve specific data.

3. **Web Frontend:** This component enhances the existing Yunikorn Web interface by providing additional features that utilize
   the historical data stored by UHS. It is loaded on the application page of Yunikorn Web.
   More details on the web component is available [here](https://github.com/G-Research/unicorn-history-server/tree/main/web).

By integrating these components, UHS provides a comprehensive view of the historical operations of a Yunikorn-managed cluster,
enabling detailed analysis and insights.
