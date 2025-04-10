# Data Center Ecology Map (EcoGrid Tycoon)

An interactive web-based strategy game that challenges players to build an environmentally sustainable global data center network. Balance economic success with ecological responsibility in this educational simulation.

![EcoGrid Tycoon](https://via.placeholder.com/800x400?text=EcoGrid+Tycoon+Screenshot)

## 🌍 Project Overview

Data Center Tycoon (EcoGrid Tycoon) is a simulation game that addresses the growing environmental impact of digital infrastructure. With data centers consuming approximately 3% of global energy and contributing significantly to carbon emissions, this project aims to educate users about sustainable technology decisions through engaging gameplay.

The game challenges players to:
- Build and manage a global network of data centers
- Balance economic constraints with environmental impact
- Make strategic decisions about location, technology, and resources
- Learn about real-world environmental factors affecting data center sustainability

## 🎮 Core Features

### Interactive Global Map
- Displays potential and existing data center locations
- Real-time marker updates as players make decisions
- Geographic visualization of environmental factors

### Environmental Assessment System
Each location is evaluated using four key metrics:
- Climate Suitability
- Renewable Energy Potential
- Grid Cleanliness
- Disaster Safety

### Tiered Construction Options
- Standard Data Center (40% efficiency)
- Eco-Optimized Center (60% efficiency)
- Next-Gen Sustainable Facility (80% efficiency)

### Economic Simulation
- Budget management starting at $10,000,000
- Carbon footprint tracking
- Revenue generation based on data center performance
- Cost-benefit analysis of environmental investments

## 🔧 Technology Stack

### Frontend
- **Framework**: React.js
- **Map Visualization**: Leaflet.js with custom dark-themed CartoDB tiles
- **Styling**: Bootstrap with custom black and gold theme
- **UI/UX**: Modern, responsive design

### Backend
- **Language**: Go (Golang)
- **Authentication**: Custom session-based authentication
- **Data Storage**: CSV files (with planned migration to PostgreSQL/MongoDB)

### Environmental Algorithms
- Research-oriented scoring system
- Simulates climate conditions, renewable potential, grid emissions, and disaster risk
- Compound environmental effects (heat island, water competition)

## 📂 Project Structure

```
data-center-ecology-map/
│
├── backend/                # Go backend
│   ├── main.go             # Main server logic
│   ├── handlers.go         # API endpoint handlers
│   └── data/               # CSV data files
│
├── frontend/               # React frontend
│   ├── src/
│   │   ├── components/     # React components
│   │   │   ├── Game.js     # Main game component
│   │   │   ├── Login.js    # Authentication
│   │   │   └── ...
│   │   ├── services/       # API service
│   │   └── App.js          # Main application
│
└── README.md               # Project documentation
```

## 🔬 Research-Based Environmental Modeling

The game incorporates advanced environmental impact models based on scientific research:

- **PUE Calculation**: Based on ASHRAE thermal guidelines and Lawrence Berkeley National Laboratory datacenter energy models
- **Carbon Emissions**: Utilizes EPA eGRID data for regional grid intensity
- **Water Impact**: Applies Water Footprint Assessment methodology (Hoekstra et al.)
- **Temperature Impact**: Based on research from "Thermal Output Impact of Data Centers" (Wang et al., 2019)
- **Density Effects**: Models compounding effects of datacenter clustering on heat and resource competition

## 🚀 Getting Started

### Prerequisites
- Node.js (v14+)
- Go (v1.16+)
- Git

### Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/data-center-ecology-map.git
cd data-center-ecology-map
```

2. Install backend dependencies
```bash
cd backend
go mod download
```

3. Install frontend dependencies
```bash
cd frontend
npm install
```

4. Start the backend server
```bash
cd backend
go run main.go
```

5. Start the frontend development server
```bash
cd datacenter_ecology_frontend
npm start
```

6. Open your browser and navigate to `http://localhost:3000`

## 🔮 Future Roadmap

- **Enhanced AI Integration**: Predictive models for climate impact and energy demand
- **Expanded Economic Modeling**: Tax incentives, operational costs, and market simulation
- **User Customization**: More dynamic gameplay options and personalization

## 📚 Educational & Real-World Impact

### Sustainability Education
- Showcases the complexities of sustainable infrastructure planning
- Demonstrates the environmental impact of technology decisions
- Encourages data-driven resource allocation and risk management

### Industry Relevance
- Prepares planners for real-world challenges in the tech industry
- Emphasizes responsible resource usage and strategic thinking
- Inspires innovative solutions for a more sustainable digital future

## 📊 Resources

- [Data Center Map](https://www.datacentermap.com/)
- [International Energy Agency (IEA)](https://www.iea.org/)
- [US Department of Energy (DOE)](https://www.energy.gov/)
- [Data Center Knowledge](https://www.datacenterknowledge.com/)

## 👥 Contributing

Contributions are welcome! Please follow these guidelines:
- Follow clean code principles
- Maintain comprehensive documentation
- Write unit and integration tests
- Adhere to sustainable coding practices

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 📧 Contact

For more information, collaboration, or support, please contact the project maintainers:
- Email: sam.karkada@gmail.com, bharathbr12@gmail.com

---

*This documentation is a living document and will be updated as the project evolves.*
