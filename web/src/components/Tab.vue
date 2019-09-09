<script>
import TabContent from 'components/TabContent.vue';
import TabLink from 'components/TabLink.vue';

export default {
    functional: true,
    components: {TabContent,TabLink},
    props: {
        name: {
            type: String,
            required: true,
        },
        selected: {
            type: Boolean,
            default: false,
        },
    },
    render: function(createElement, context) {
        const cleanedName = context.props.name.toLowerCase().replace(/\W/g, '-');
        const props = {
            id: 'tab-' + cleanedName,
            href: '#' + cleanedName,
            ...context.props,
        };
        return [
            createElement(TabLink, {props}),
            createElement(TabContent, {props}, context.children),
        ];
    },
};
</script>
