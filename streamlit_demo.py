import streamlit as st
import networkx as nx
import matplotlib.pyplot as plt
import yaml
import json

class DockerComposeVisualizer:
    def __init__(self, compose_data):
        self.graph = self.parse_compose_data(compose_data)

    def parse_compose_data(self, compose_data):
        # Parse the docker-compose data to create a directed graph
        graph = nx.DiGraph()
        services = compose_data.get('services', {})

        for service_name, service in services.items():
            graph.add_node(service_name)
            depends_on = service.get('depends_on', [])
            
            # Handle both list and dict formats for depends_on
            if isinstance(depends_on, dict):
                depends_on = list(depends_on.keys())
            
            for dependency in depends_on:
                if isinstance(dependency, dict):
                    dependency = list(dependency.keys())[0]
                graph.add_edge(service_name, dependency)

        return graph

    def topological_sort(self):
        try:
            return list(nx.topological_sort(self.graph))
        except nx.NetworkXError:
            return None

    def detect_cycles(self):
        return list(nx.simple_cycles(self.graph))

    def visualize_graph(self):
        fig, ax = plt.subplots(figsize=(12, 8))
        pos = nx.spring_layout(self.graph, k=2, iterations=50)
        nx.draw_networkx_nodes(self.graph, pos, node_color='lightblue', node_size=1500, ax=ax)
        nx.draw_networkx_labels(self.graph, pos, font_size=10, ax=ax)
        nx.draw_networkx_edges(self.graph, pos, edge_color='gray', arrows=True, 
                               arrowsize=20, arrowstyle='->', ax=ax, connectionstyle='arc3,rad=0.1')
        ax.set_title('Docker Compose Dependency Graph', fontsize=16)
        ax.axis('off')
        return fig

def main():
    st.set_page_config(page_title='DockSmith Visualizer', layout='wide')
    
    st.title('🐳 DockSmith - Docker Compose Visualizer')
    st.write('Visualize Docker Compose dependency graphs with topological sorting and cycle detection.')
    
    # Sidebar for options
    st.sidebar.header('Options')
    
    # File upload section
    col1, col2 = st.columns(2)
    
    with col1:
        st.subheader('Upload Docker Compose File')
        uploaded_file = st.file_uploader('Choose a docker-compose.yml file', type=['yml', 'yaml'])
    
    compose_data = None
    
    if uploaded_file is not None:
        try:
            compose_data = yaml.safe_load(uploaded_file)
            st.success('✅ File loaded successfully!')
        except Exception as e:
            st.error(f'❌ Error loading file: {e}')
            return
    
    # Example compose data
    with col2:
        st.subheader('Or Use Example')
        if st.button('Load Example Docker Compose'):
            compose_data = {
                'services': {
                    'web': {
                        'depends_on': ['db', 'cache']
                    },
                    'db': {
                        'depends_on': ['cache']
                    },
                    'cache': {},
                    'api': {
                        'depends_on': ['web']
                    }
                }
            }
            st.success('✅ Example loaded!')
    
    if compose_data:
        visualizer = DockerComposeVisualizer(compose_data)
        
        # Display graph
        st.subheader('📊 Dependency Graph')
        fig = visualizer.visualize_graph()
        st.pyplot(fig)
        
        # Topological sorting
        st.subheader('📈 Topological Sorting (Build Order)')
        sorted_services = visualizer.topological_sort()
        if sorted_services:
            st.success('✅ No cycles detected!')
            st.write('**Recommended build order:**')
            for i, service in enumerate(sorted_services, 1):
                st.write(f'{i}. {service}')
        else:
            st.error('❌ Cycles detected! Cannot determine build order.')
        
        # Cycle detection
        st.subheader('🔄 Cycle Detection')
        cycles = visualizer.detect_cycles()
        if cycles:
            st.error(f'❌ {len(cycles)} cycle(s) detected:')
            for cycle in cycles:
                st.write(f'  → {" → ".join(cycle)} → {cycle[0]}')
        else:
            st.success('✅ No cycles detected. Graph is acyclic!')
        
        # Graph statistics
        st.subheader('📋 Graph Statistics')
        col1, col2, col3 = st.columns(3)
        with col1:
            st.metric('Services', len(visualizer.graph.nodes()))
        with col2:
            st.metric('Dependencies', len(visualizer.graph.edges()))
        with col3:
            st.metric('Cycles', len(cycles))

if __name__ == '__main__':
    main()
